package vertexai

import (
	"net/http"
	"one-api/common"
	"one-api/common/requester"
	"one-api/providers/gemini"
	"one-api/providers/vertexai/category"
	"one-api/types"
)

func (p *VertexAIProvider) CreateGeminiChat(request *gemini.GeminiChatRequest) (*gemini.GeminiChatResponse, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getGeminiRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	geminiResponse := &gemini.GeminiChatResponse{}
	// 发送请求
	_, openaiErr := p.Requester.SendRequest(req, geminiResponse, false)
	if openaiErr != nil {
		return nil, openaiErr
	}

	if len(geminiResponse.Candidates) == 0 {
		return nil, common.StringErrorWrapper("no candidates", "no_candidates", http.StatusInternalServerError)
	}

	usage := p.GetUsage()
	*usage = convertOpenAIUsage(geminiResponse.UsageMetadata)

	return geminiResponse, nil
}

func (p *VertexAIProvider) CreateGeminiChatStream(request *gemini.GeminiChatRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getGeminiRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	chatHandler := &gemini.GeminiRelayStreamHandler{
		Usage:     p.Usage,
		ModelName: request.Model,
		Prefix:    `data: `,
	}

	// 发送请求
	resp, openaiErr := p.Requester.SendRequestRaw(req)
	if openaiErr != nil {
		return nil, openaiErr
	}

	stream, openaiErr := requester.RequestNoTrimStream(p.Requester, resp, chatHandler.HandlerStream)
	if openaiErr != nil {
		return nil, openaiErr
	}

	return stream, nil
}

func (p *VertexAIProvider) getGeminiRequest(request *gemini.GeminiChatRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	var err error
	p.Category, err = category.GetCategory(request.Model)
	if err != nil || p.Category.ChatComplete == nil || p.Category.ResponseChatComplete == nil {
		return nil, common.StringErrorWrapperLocal("vertexAI gemini provider not found", "vertexAI_err", http.StatusInternalServerError)
	}

	otherUrl := p.Category.GetOtherUrl(request.Stream)
	modelName := p.Category.GetModelName(request.Model)

	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(modelName, otherUrl)
	if fullRequestURL == "" {
		return nil, common.ErrorWrapperLocal(nil, "invalid_vertexai_config", http.StatusInternalServerError)
	}

	headers := p.GetRequestHeaders()
	if headers == nil {
		return nil, common.ErrorWrapperLocal(nil, "invalid_vertexai_config", http.StatusInternalServerError)
	}

	// 错误处理
	p.Requester.ErrorHandler = RequestErrorHandle(p.Category.ErrorHandler)

	// 使用BaseProvider的统一方法创建请求，支持额外参数处理
	req, errWithCode := p.NewRequestWithCustomParams(http.MethodPost, fullRequestURL, request.GetJsonRaw(), headers, request.Model)
	if errWithCode != nil {
		return nil, errWithCode
	}
	return req, nil
}

func convertOpenAIUsage(geminiUsage *gemini.GeminiUsageMetadata) types.Usage {
	if geminiUsage == nil {
		return types.Usage{}
	}
	return types.Usage{
		PromptTokens:     geminiUsage.PromptTokenCount,
		CompletionTokens: geminiUsage.CandidatesTokenCount + geminiUsage.ThoughtsTokenCount,
		TotalTokens:      geminiUsage.TotalTokenCount,

		CompletionTokensDetails: types.CompletionTokensDetails{
			ReasoningTokens: geminiUsage.ThoughtsTokenCount,
		},
	}
}
