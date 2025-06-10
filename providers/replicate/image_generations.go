package replicate

import (
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/types"
	"time"
)

func (p *ReplicateProvider) CreateImageGenerations(request *types.ImageRequest) (*types.ImageResponse, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(config.RelayModeImagesGenerations)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url, request.Model)
	if fullRequestURL == "" {
		return nil, common.ErrorWrapper(nil, "invalid_recraft_config", http.StatusInternalServerError)
	}

	// 获取请求头
	headers := p.GetRequestHeaders()

	replicateRequest := convertFromIamgeOpenai(request)
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(replicateRequest), p.Requester.WithHeader(headers))

	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	replicateResponse := &ReplicateResponse[string]{}

	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, replicateResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	replicateResponse, err = getPrediction(p, replicateResponse)
	if err != nil {
		return nil, common.ErrorWrapper(err, "prediction_failed", http.StatusInternalServerError)
	}

	if replicateResponse.Output == "" {
		replicateResponse.Output = replicateResponse.Urls.Stream
	}

	// PromptTokens保持之前根据prompt内容计算的值
	// CompletionTokens根据生成的图像数量计算，避免空回复计费问题
	imageCount := 1 // Replicate通常生成一张图
	p.Usage.CompletionTokens = imageCount * 258
	p.Usage.TotalTokens = p.Usage.PromptTokens + p.Usage.CompletionTokens

	return p.convertToImageOpenai(replicateResponse)
}

func convertFromIamgeOpenai(request *types.ImageRequest) *ReplicateRequest[ReplicateImageRequest] {
	replicateRequest := &ReplicateRequest[ReplicateImageRequest]{
		Input: ReplicateImageRequest{
			Prompt:           request.Prompt,
			OutputFormat:     request.ResponseFormat,
			Size:             request.Size,
			AspectRatio:      request.AspectRatio,
			OutputQuality:    request.OutputQuality,
			SafetyTolerance:  request.SafetyTolerance,
			PromptUpsampling: request.PromptUpsampling,
		},
	}
	return replicateRequest
}

func (p *ReplicateProvider) convertToImageOpenai(response *ReplicateResponse[string]) (*types.ImageResponse, *types.OpenAIErrorWithStatusCode) {
	openaiResponse := &types.ImageResponse{
		Created: time.Now().Unix(),
		Data: []types.ImageResponseDataInner{
			{
				URL: response.Output,
			},
		},
	}

	return openaiResponse, nil
}
