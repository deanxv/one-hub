package relay

import (
	"errors"
	"net/http"
	"one-api/common"
	providersBase "one-api/providers/base"
	"one-api/types"
	"strings"

	"github.com/gin-gonic/gin"
)

type relayImageGenerations struct {
	relayBase
	request types.ImageRequest
}

func newRelayImageGenerations(c *gin.Context) *relayImageGenerations {
	relay := &relayImageGenerations{}
	relay.c = c
	return relay
}

func (r *relayImageGenerations) setRequest() error {
	// 检查是否是Gemini格式的请求 (model:predict)
	path := r.c.Request.URL.Path
	if strings.Contains(path, ":predict") {
		return r.setGeminiRequest()
	}

	// 标准OpenAI格式
	if err := common.UnmarshalBodyReusable(r.c, &r.request); err != nil {
		return err
	}

	if r.request.Model == "" {
		r.request.Model = "dall-e-2"
	}

	if r.request.N == 0 {
		r.request.N = 1
	}

	if strings.HasPrefix(r.request.Model, "dall-e") {
		if r.request.Size == "" {
			r.request.Size = "1024x1024"
		}

		if r.request.Quality == "" {
			r.request.Quality = "standard"
		}
	}

	r.setOriginalModel(r.request.Model)

	return nil
}

// Gemini格式请求处理
func (r *relayImageGenerations) setGeminiRequest() error {
	// 直接从Gin的路由参数获取模型名（包含冒号部分）
	modelParam := r.c.Param("model")
	if modelParam == "" {
		return errors.New("model parameter not found")
	}

	// 分离模型名和动作 (imagen-3.0-generate-002:predict -> imagen-3.0-generate-002)
	modelName := strings.Split(modelParam, ":")[0]

	// 解析Gemini格式的请求体 - 使用通用结构以支持参数透传
	var geminiRequest struct {
		Instances []struct {
			Prompt string `json:"prompt"`
		} `json:"instances"`
		Parameters map[string]interface{} `json:"parameters"`
	}

	if err := common.UnmarshalBodyReusable(r.c, &geminiRequest); err != nil {
		return err
	}

	// 转换为标准格式
	if len(geminiRequest.Instances) == 0 {
		return errors.New("no instances provided")
	}

	r.request = types.ImageRequest{
		Model:       modelName,
		Prompt:      geminiRequest.Instances[0].Prompt,
		N:           1, // 默认值
		ExtraParams: make(map[string]interface{}),
	}

	// 处理parameters中的所有参数
	for key, value := range geminiRequest.Parameters {
		switch key {
		case "sampleCount":
			if sampleCount, ok := value.(float64); ok {
				r.request.N = int(sampleCount)
			}
		case "aspectRatio":
			if aspectRatio, ok := value.(string); ok && aspectRatio != "" {
				r.request.AspectRatio = &aspectRatio
			}
		default:
			// 其他所有参数都作为额外参数透传
			r.request.ExtraParams[key] = value
		}
	}

	if r.request.N == 0 {
		r.request.N = 1
	}

	r.setOriginalModel(r.request.Model)

	return nil
}

func (r *relayImageGenerations) getPromptTokens() (int, error) {
	// PromptTokens应该根据请求中的prompt文本计算，而不是图像参数
	return common.CountTokenText(r.request.Prompt, r.getOriginalModel()), nil
}

func (r *relayImageGenerations) send() (err *types.OpenAIErrorWithStatusCode, done bool) {
	provider, ok := r.provider.(providersBase.ImageGenerationsInterface)
	if !ok {
		err = common.StringErrorWrapperLocal("channel not implemented", "channel_error", http.StatusServiceUnavailable)
		done = true
		return
	}

	r.request.Model = r.modelName

	response, err := provider.CreateImageGenerations(&r.request)
	if err != nil {
		return
	}
	err = responseJsonClient(r.c, response)

	if err != nil {
		done = true
	}

	return
}
