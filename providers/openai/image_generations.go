package openai

import (
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/types"
)

func (p *OpenAIProvider) CreateImageGenerations(request *types.ImageRequest) (*types.ImageResponse, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.GetRequestTextBody(config.RelayModeImagesGenerations, request.Model, request)
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

	// 检测是否错误
	openaiErr := ErrorHandle(&response.OpenAIErrorResponse)
	if openaiErr != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *openaiErr,
			StatusCode:  http.StatusBadRequest,
		}
		return nil, errWithCode
	}

	if response.Usage != nil && response.Usage.TotalTokens > 0 {
		*p.Usage = *response.Usage.ToOpenAIUsage()
	} else {
		// 如果没有返回usage信息，计算生图的CompletionTokens
		imageCount := len(response.Data)
		// PromptTokens保持之前根据prompt内容计算的值
		// CompletionTokens根据生成的图像数量计算，避免空回复计费问题
		p.Usage.CompletionTokens = imageCount * 258
		p.Usage.TotalTokens = p.Usage.PromptTokens + p.Usage.CompletionTokens
	}

	return &response.ImageResponse, nil

}

func IsWithinRange(element string, value int) bool {
	if _, ok := common.DalleGenerationImageAmounts[element]; !ok {
		return true
	}
	minCount := common.DalleGenerationImageAmounts[element][0]
	maxCount := common.DalleGenerationImageAmounts[element][1]

	return value >= minCount && value <= maxCount
}
