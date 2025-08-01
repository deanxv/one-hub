package recraftAI

import (
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/types"
)

func (p *RecraftProvider) CreateImageGenerations(request *types.ImageRequest) (*types.ImageResponse, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(config.RelayModeImagesGenerations)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url)
	if fullRequestURL == "" {
		return nil, common.ErrorWrapper(nil, "invalid_recraft_config", http.StatusInternalServerError)
	}

	// 获取请求头
	headers := p.GetRequestHeaders()
	body, exists := p.GetRawBody()
	if !exists {
		return nil, common.StringErrorWrapperLocal("request body not found", "request_body_not_found", http.StatusInternalServerError)
	}

	req, err := p.Requester.NewRequest(
		http.MethodPost,
		fullRequestURL,
		p.Requester.WithBody(body),
		p.Requester.WithHeader(headers),
		p.Requester.WithContentType(p.Context.Request.Header.Get("Content-Type")))
	req.ContentLength = p.Context.Request.ContentLength

	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	recraftResponse := &types.ImageResponse{}

	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, recraftResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	imageCount := len(recraftResponse.Data)
	// PromptTokens保持之前根据prompt内容计算的值
	// CompletionTokens根据生成的图像数量计算，避免空回复计费问题
	p.Usage.CompletionTokens = imageCount * 258
	p.Usage.TotalTokens = p.Usage.PromptTokens + p.Usage.CompletionTokens

	return recraftResponse, nil
}
