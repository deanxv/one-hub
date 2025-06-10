package openai

import (
	"net/http"
	"one-api/common/config"
	"one-api/types"
)

func (p *OpenAIProvider) CreateImageVariations(request *types.ImageEditRequest) (*types.ImageResponse, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getRequestImageBody(config.RelayModeImagesVariations, request.Model, request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	response := &OpenAIProviderImageResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, response, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	openaiErr := ErrorHandle(&response.OpenAIErrorResponse)
	if openaiErr != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *openaiErr,
			StatusCode:  http.StatusBadRequest,
		}
		return nil, errWithCode
	}

	// 计算图像变换的CompletionTokens
	imageCount := len(response.Data)
	// PromptTokens保持之前根据prompt内容计算的值
	// CompletionTokens根据生成的图像数量计算，避免空回复计费问题
	p.Usage.CompletionTokens = imageCount * 258
	p.Usage.TotalTokens = p.Usage.PromptTokens + p.Usage.CompletionTokens

	return &response.ImageResponse, nil
}
