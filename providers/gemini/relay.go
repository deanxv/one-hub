package gemini

import (
	"bytes"
	"encoding/json"
	"net/http"
	"one-api/common"
	"one-api/common/requester"
	"one-api/types"
	"strings"
)

type GeminiRelayStreamHandler struct {
	Usage     *types.Usage
	Prefix    string
	ModelName string

	Key string
}

func (p *GeminiProvider) CreateGeminiChat(request *GeminiChatRequest) (*GeminiChatResponse, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getChatRequest(request, true)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	geminiResponse := &GeminiChatResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, geminiResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 只有非 countTokens 请求才检查 candidates
	if request.Action != "countTokens" && len(geminiResponse.Candidates) == 0 {
		return nil, common.StringErrorWrapper("no candidates", "no_candidates", http.StatusInternalServerError)
	}

	usage := p.GetUsage()
	*usage = ConvertOpenAIUsage(geminiResponse.UsageMetadata)

	return geminiResponse, nil
}

func (p *GeminiProvider) CreateGeminiChatStream(request *GeminiChatRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getChatRequest(request, true)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	channel := p.GetChannel()

	chatHandler := &GeminiRelayStreamHandler{
		Usage:     p.Usage,
		ModelName: request.Model,
		Prefix:    `data: `,

		Key: channel.Key,
	}

	// 发送请求
	resp, errWithCode := p.Requester.SendRequestRaw(req)
	if errWithCode != nil {
		return nil, errWithCode
	}

	stream, errWithCode := requester.RequestNoTrimStream(p.Requester, resp, chatHandler.HandlerStream)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return stream, nil
}

func (h *GeminiRelayStreamHandler) HandlerStream(rawLine *[]byte, dataChan chan string, errChan chan error) {
	rawStr := string(*rawLine)
	// 如果rawLine 前缀不为data:，则直接返回
	if !strings.HasPrefix(rawStr, h.Prefix) {
		dataChan <- rawStr
		return
	}

	noSpaceLine := bytes.TrimSpace(*rawLine)
	noSpaceLine = noSpaceLine[6:]

	var geminiResponse GeminiChatResponse
	err := json.Unmarshal(noSpaceLine, &geminiResponse)
	if err != nil {
		errChan <- ErrorToGeminiErr(err)
		return
	}

	if geminiResponse.ErrorInfo != nil {
		cleaningError(geminiResponse.ErrorInfo, h.Key)
		errChan <- geminiResponse.ErrorInfo
		return
	}

	if geminiResponse.UsageMetadata == nil {
		dataChan <- rawStr
		return
	}

	h.Usage.PromptTokens = geminiResponse.UsageMetadata.PromptTokenCount

	// 计算 completion tokens，确保不为负数
	completionTokens := geminiResponse.UsageMetadata.CandidatesTokenCount + geminiResponse.UsageMetadata.ThoughtsTokenCount
	if completionTokens < 0 {
		completionTokens = 0
	}
	h.Usage.CompletionTokens = completionTokens
	h.Usage.CompletionTokensDetails.ReasoningTokens = geminiResponse.UsageMetadata.ThoughtsTokenCount

	// 如果 TotalTokenCount 为 0 但有 PromptTokenCount，则计算总数
	totalTokens := geminiResponse.UsageMetadata.TotalTokenCount
	if totalTokens == 0 && geminiResponse.UsageMetadata.PromptTokenCount > 0 {
		totalTokens = geminiResponse.UsageMetadata.PromptTokenCount + completionTokens
	}
	h.Usage.TotalTokens = totalTokens

	dataChan <- rawStr
}
