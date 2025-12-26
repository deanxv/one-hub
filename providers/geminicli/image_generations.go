package geminicli

import (
	"net/http"
	"one-api/common"
	"one-api/common/utils"
	"one-api/providers/gemini"
	"one-api/types"
)

// CreateImageGenerations 创建图像生成（使用Gemini原生格式）
func (p *GeminiCliProvider) CreateImageGenerations(request *types.ImageRequest) (*types.ImageResponse, *types.OpenAIErrorWithStatusCode) {
	imageConfig := &gemini.ImageConfig{}
	if request.AspectRatio != nil {
		imageConfig.AspectRatio = *request.AspectRatio
	} else {
		switch request.Size {
		case "1024x1792":
			imageConfig.AspectRatio = "9:16"
		case "1792x1024":
			imageConfig.AspectRatio = "16:9"
		default:
			imageConfig.AspectRatio = "1:1"
		}
	}

	geminiRequest := &gemini.GeminiChatRequest{
		Model: request.Model,
		Contents: []gemini.GeminiChatContent{
			{
				Role: "user",
				Parts: []gemini.GeminiPart{
					{
						Text: request.Prompt,
					},
				},
			},
		},
		GenerationConfig: gemini.GeminiChatGenerationConfig{
			ImageConfig: imageConfig,
		},
	}

	// 构建内部API请求
	req, errWithCode := p.getImageRequest(geminiRequest)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	// 使用包装的响应结构
	cliResponse := &GeminiCliResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, cliResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 提取实际的Gemini响应
	if cliResponse.Response == nil {
		return nil, common.StringErrorWrapper("no response in upstream response", "no_response", http.StatusInternalServerError)
	}

	geminiResponse := cliResponse.Response

	if len(geminiResponse.Candidates) == 0 {
		return nil, common.StringErrorWrapper("no candidates in response", "no_candidates", http.StatusInternalServerError)
	}

	// 从响应中提取图像数据
	openaiResponse := &types.ImageResponse{
		Created: utils.GetTimestamp(),
		Data:    make([]types.ImageResponseDataInner, 0),
	}

	for _, candidate := range geminiResponse.Candidates {
		if candidate.Content.Parts == nil {
			continue
		}

		for _, part := range candidate.Content.Parts {
			if part.InlineData != nil && part.InlineData.Data != "" {
				openaiResponse.Data = append(openaiResponse.Data, types.ImageResponseDataInner{
					B64JSON: part.InlineData.Data,
				})
			}
		}
	}

	imageCount := len(openaiResponse.Data)
	// 如果imageCount为0，则返回错误
	if imageCount == 0 {
		return nil, common.StringErrorWrapper("no image generated", "no_image_generated", http.StatusInternalServerError)
	}

	usage := p.GetUsage()
	if geminiResponse.UsageMetadata != nil {
		*usage = gemini.ConvertOpenAIUsage(geminiResponse.UsageMetadata)
	} else {
		// PromptTokens保持之前根据prompt内容计算的值
		// CompletionTokens根据生成的图像数量计算，避免空回复计费问题
		usage.CompletionTokens = imageCount * 258
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	return openaiResponse, nil
}

// getImageRequest 构建内部API请求
func (p *GeminiCliProvider) getImageRequest(geminiRequest *gemini.GeminiChatRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	action := "generateContent"
	fullRequestURL := p.GetFullRequestURL(action, geminiRequest.Model)

	// 获取请求头
	headers, err := p.getRequestHeadersInternal()
	if err != nil {
		return nil, common.StringErrorWrapper(err.Error(), "geminicli_token_error", http.StatusUnauthorized)
	}

	requestBody := &GeminiCliRequest{
		Model:   geminiRequest.Model,
		Project: p.ProjectID,
		Request: geminiRequest,
	}
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(requestBody), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "create_request_failed", http.StatusInternalServerError)
	}

	return req, nil
}
