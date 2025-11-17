package relay

import (
	"encoding/json"
	"errors"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/providers/gemini"
	"one-api/safty"
	"one-api/types"
	"strings"

	"github.com/gin-gonic/gin"
)

type relayVeoOnly struct {
	relayBase
	veoRequest *gemini.VeoVideoRequest
}

func NewRelayVeoOnly(c *gin.Context) *relayVeoOnly {
	c.Set("allow_channel_type", AllowGeminiChannelType)

	relay := &relayVeoOnly{
		relayBase: relayBase{
			allowHeartbeat: true,
			c:              c,
		},
	}

	return relay
}

func (r *relayVeoOnly) setRequest() error {
	// 处理 predictLongRunning 请求
	modelAction := r.c.Param("model")
	if modelAction == "" {
		return errors.New("model is required")
	}

	modelList := strings.Split(modelAction, ":")
	if len(modelList) != 2 {
		return errors.New("model error")
	}

	action := modelList[1]
	if action != "predictLongRunning" {
		return errors.New("unsupported action for Veo")
	}

	// 解析 Veo 请求体
	r.veoRequest = &gemini.VeoVideoRequest{}
	if err := common.UnmarshalBodyReusable(r.c, r.veoRequest); err != nil {
		return err
	}

	// 验证请求
	if len(r.veoRequest.Instances) == 0 {
		return errors.New("instances is required")
	}

	if len(r.veoRequest.Instances) > 1 {
		return errors.New("only one instance is supported, multiple video generation is not allowed")
	}

	for _, instance := range r.veoRequest.Instances {
		if instance.Prompt == "" {
			return errors.New("prompt is required in instances")
		}
	}

	r.setOriginalModel(modelList[0])
	r.c.Set("original_model", modelList[0])

	return nil
}

func (r *relayVeoOnly) getRequest() interface{} {
	return r.veoRequest
}

func (r *relayVeoOnly) IsStream() bool {
	return false // Veo 3.0 返回视频文件，不是流式响应
}

func (r *relayVeoOnly) getPromptTokens() (int, error) {
	// 对于 Veo 3.0，只计算 prompt 的 token 数量
	totalTokens := 0
	for _, instance := range r.veoRequest.Instances {
		tokens := common.CountTokenText(instance.Prompt, r.getModelName())
		totalTokens += tokens
	}

	return totalTokens, nil
}

func (r *relayVeoOnly) send() (err *types.OpenAIErrorWithStatusCode, done bool) {
	geminiProvider, ok := r.provider.(*gemini.GeminiProvider)
	if !ok {
		return common.StringErrorWrapperLocal("channel not implemented", "channel_error", http.StatusServiceUnavailable), true
	}

	// 内容审查
	if config.EnableSafe {
		for _, instance := range r.veoRequest.Instances {
			if instance.Prompt != "" {
				CheckResult, _ := safty.CheckContent(instance.Prompt)
				if !CheckResult.IsSafe {
					err = common.StringErrorWrapperLocal(CheckResult.Reason, CheckResult.Code, http.StatusBadRequest)
					done = true
					return
				}
			}
		}
	}

	// 处理视频生成请求（包含轮询和下载）
	// 重要：使用同一个 geminiProvider 实例，确保整个流程使用相同的 API Key
	videoData, contentType, errWithCode := geminiProvider.CreateVeoVideoAndDownload(r.veoRequest, r.modelName)
	if errWithCode != nil {
		return errWithCode, false
	}

	if r.heartbeat != nil {
		r.heartbeat.Stop()
	}

	// 更新使用统计 - 视频生成按固定 token 计费
	usage := r.provider.GetUsage()
	if usage != nil {
		// 对于 Veo 3.0，设置固定的输出 token（因为返回的是视频文件）
		usage.CompletionTokens = 1000 // 视频生成固定费用
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	// 设置响应头并返回视频文件
	r.c.Header("Content-Type", contentType)
	r.c.Header("Content-Disposition", "attachment; filename=generated_video.mp4")
	r.c.Data(http.StatusOK, contentType, videoData)

	return nil, true
}

// GetError 实现错误处理，遵循项目的错误处理模式
func (r *relayVeoOnly) GetError(err *types.OpenAIErrorWithStatusCode) (int, any) {
	newErr := FilterOpenAIErr(r.c, err)

	// 将错误转换为 Gemini 格式，因为 Veo 3.0 是 Gemini 的一部分
	geminiErr := gemini.OpenaiErrToGeminiErr(&newErr)

	return newErr.StatusCode, geminiErr.GeminiErrorResponse
}

// HandleJsonError 处理 JSON 错误响应
func (r *relayVeoOnly) HandleJsonError(err *types.OpenAIErrorWithStatusCode) {
	statusCode, response := r.GetError(err)
	r.c.JSON(statusCode, response)
}

// HandleStreamError 处理流式错误响应（虽然 Veo 不支持流式，但需要实现接口）
func (r *relayVeoOnly) HandleStreamError(err *types.OpenAIErrorWithStatusCode) {
	_, response := r.GetError(err)

	str, jsonErr := json.Marshal(response)
	if jsonErr != nil {
		return
	}
	r.c.Writer.Write([]byte("data: " + string(str) + "\n\n"))
	r.c.Writer.Flush()
}
