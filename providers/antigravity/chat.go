package antigravity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/common/logger"
	"one-api/common/requester"
	"one-api/common/utils"
	"one-api/providers/base"
	"one-api/providers/gemini"
	"one-api/types"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateChatCompletion 创建聊天补全（非流式）
func (p *AntigravityProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	// 转换为Gemini格式
	geminiRequest, errWithCode := gemini.ConvertFromChatOpenai(request)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 修复空 Parameters 问题：Claude API 要求 input_schema 必须存在
	fixNilToolParameters(geminiRequest)

	// 构建内部API请求
	req, errWithCode := p.getChatRequest(geminiRequest, false, false)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	// 使用包装的响应结构
	antigravityResponse := &AntigravityResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, antigravityResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 提取实际的 Gemini 响应
	if antigravityResponse.Response == nil {
		return nil, common.StringErrorWrapper("no response in upstream response", "no_response", http.StatusInternalServerError)
	}

	return gemini.ConvertToChatOpenai(p, antigravityResponse.Response, request)
}

// CreateChatCompletionStream 创建聊天补全（流式）
func (p *AntigravityProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
	// 转换为Gemini格式
	geminiRequest, errWithCode := gemini.ConvertFromChatOpenai(request)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 修复空 Parameters 问题：Claude API 要求 input_schema 必须存在
	fixNilToolParameters(geminiRequest)

	// 构建内部API请求
	req, errWithCode := p.getChatRequest(geminiRequest, true, false)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	// 发送请求
	resp, errWithCode := p.Requester.SendRequestRaw(req)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 使用 Antigravity 专用的流处理器
	chatHandler := &AntigravityStreamHandler{
		Usage:   p.Usage,
		Request: request,
		Context: p.Context,
	}

	return requester.RequestStream(p.Requester, resp, chatHandler.HandlerStream)
}

// fixNilToolParameters 修复空的 tool parameters
// gemini.ConvertFromChatOpenai 会将空 properties 的 schema 设为 nil
// 但 Claude API 要求 input_schema 必须存在
func fixNilToolParameters(geminiRequest *gemini.GeminiChatRequest) {
	if geminiRequest == nil {
		return
	}

	for i := range geminiRequest.Tools {
		for j := range geminiRequest.Tools[i].FunctionDeclarations {
			if geminiRequest.Tools[i].FunctionDeclarations[j].Parameters == nil {
				// 每次创建新的 map 避免共享引用
				geminiRequest.Tools[i].FunctionDeclarations[j].Parameters = map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				}
			}
		}
	}
}

// generateRequestID 生成请求 ID
func generateRequestID() string {
	return fmt.Sprintf("req-%s", uuid.New().String())
}

// getChatRequest 构建内部API请求
func (p *AntigravityProvider) getChatRequest(geminiRequest *gemini.GeminiChatRequest, isStream bool, isRelay bool) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	// 确定请求URL
	action := "generateContent"
	if isStream {
		action = "streamGenerateContent"
	}

	fullRequestURL := p.GetFullRequestURL(action, geminiRequest.Model)

	// 获取请求头
	headers, err := p.getRequestHeadersInternal()
	if err != nil {
		return nil, p.handleTokenError(err)
	}

	// 只有在 relay 模式下才清理数据（与 gemini provider 保持一致）
	var geminiRequestBody any
	if isRelay {
		// 使用原始请求体（避免序列化/反序列化导致数据丢失）
		rawData, exists := p.GetRawBody()
		if !exists {
			return nil, common.StringErrorWrapperLocal("request body not found", "request_body_not_found", http.StatusInternalServerError)
		}

		// 直接在原始数据上操作，避免多次序列化/反序列化
		var rawMap map[string]interface{}
		if err := json.Unmarshal(rawData, &rawMap); err != nil {
			return nil, common.ErrorWrapper(err, "unmarshal_request_failed", http.StatusInternalServerError)
		}

		// 确保 contents 中每个 content 都有 role 字段
		// 注意：不删除 functionCall/functionResponse 的 id 字段，Antigravity API 支持该字段
		if contents, ok := rawMap["contents"].([]interface{}); ok {
			for _, content := range contents {
				if contentMap, ok := content.(map[string]interface{}); ok {
					// 确保每个 content 都有 role 字段
					if _, hasRole := contentMap["role"]; !hasRole {
						contentMap["role"] = "user"
					}
				}
			}
		}

		delete(rawMap, "model")

		// 使用清理后的原始数据作为 Gemini 请求体
		geminiRequestBody = rawMap
	} else {
		geminiRequestBody = geminiRequest
	}

	if isStream {
		fullRequestURL += "?alt=sse"
	}

	// 处理额外参数
	customParams, err := p.CustomParameterHandler()
	if err != nil {
		return nil, common.ErrorWrapper(err, "custom_parameter_error", http.StatusInternalServerError)
	}

	// 如果有额外参数，将其合并到 Gemini 请求体中
	var finalGeminiRequest any = geminiRequestBody
	if customParams != nil {
		// 将 Gemini 请求体转换为 map，以便添加额外参数
		var geminiRequestMap map[string]interface{}

		// 检查 geminiRequestBody 是否已经是 map 类型
		if rawMap, ok := geminiRequestBody.(map[string]interface{}); ok {
			geminiRequestMap = rawMap
		} else {
			// 否则进行 JSON 编码
			requestBytes, err := json.Marshal(geminiRequestBody)
			if err != nil {
				return nil, common.ErrorWrapper(err, "marshal_request_failed", http.StatusInternalServerError)
			}

			err = json.Unmarshal(requestBytes, &geminiRequestMap)
			if err != nil {
				return nil, common.ErrorWrapper(err, "unmarshal_request_failed", http.StatusInternalServerError)
			}
		}

		// 处理自定义额外参数
		geminiRequestMap = p.MergeCustomParams(geminiRequestMap, customParams, geminiRequest.Model)
		finalGeminiRequest = geminiRequestMap
	}

	// 转换为 map 以便处理
	var requestMap map[string]interface{}
	if m, ok := finalGeminiRequest.(map[string]interface{}); ok {
		requestMap = m
	} else {
		requestBytes, marshalErr := json.Marshal(finalGeminiRequest)
		if marshalErr != nil {
			return nil, common.ErrorWrapper(marshalErr, "marshal_request_failed", http.StatusInternalServerError)
		}
		if unmarshalErr := json.Unmarshal(requestBytes, &requestMap); unmarshalErr != nil {
			return nil, common.ErrorWrapper(unmarshalErr, "unmarshal_request_failed", http.StatusInternalServerError)
		}
	}

	requestMap["session_id"] = fmt.Sprintf("session-%s", uuid.New().String())

	// Antigravity 不需要 safetySettings
	delete(requestMap, "safetySettings")

	applyAntigravityGenerationConfigDefaults(requestMap)
	convertToolsToAntigravityFormat(requestMap)
	applyToolConfig(requestMap)
	reorganizeToolMessages(requestMap)

	requestBody := map[string]interface{}{
		"model":     geminiRequest.Model,
		"project":   p.ProjectID,
		"requestId": generateRequestID(),
		"userAgent": "antigravity",
		"request":   requestMap,
	}

	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(requestBody), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "create_request_failed", http.StatusInternalServerError)
	}

	return req, nil
}

// AntigravityStreamHandler Antigravity 流式响应处理器
type AntigravityStreamHandler struct {
	Usage   *types.Usage
	Request *types.ChatCompletionRequest
	Context *gin.Context
}

// HandlerStream 处理流式响应
func (h *AntigravityStreamHandler) HandlerStream(rawLine *[]byte, dataChan chan string, errChan chan error) {
	rawStr := string(*rawLine)

	// 如果不是 data: 开头，直接返回
	if !strings.HasPrefix(rawStr, "data: ") {
		return
	}

	// 去除 "data: " 前缀
	noSpaceLine := bytes.TrimSpace(*rawLine)
	noSpaceLine = noSpaceLine[6:] // 去除 "data: "

	// 解析包装的响应
	var antigravityResponse AntigravityResponse
	err := json.Unmarshal(noSpaceLine, &antigravityResponse)
	if err != nil {
		logger.SysError(fmt.Sprintf("Failed to unmarshal Antigravity stream response: %s", err.Error()))
		errChan <- common.ErrorToOpenAIError(err)
		return
	}

	// 提取实际的 Gemini 响应
	if antigravityResponse.Response == nil {
		logger.SysError("Antigravity stream response has no 'response' field")
		return
	}

	geminiResponse := antigravityResponse.Response

	// 检查错误
	if geminiResponse.ErrorInfo != nil {
		errChan <- geminiResponse.ErrorInfo
		return
	}

	// 更新 usage
	if geminiResponse.UsageMetadata != nil {
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
	}

	// 转换为 OpenAI 流式响应
	h.convertToOpenaiStream(geminiResponse, dataChan)
}

// convertToOpenaiStream 将 Gemini 响应转换为 OpenAI 流式格式
func (h *AntigravityStreamHandler) convertToOpenaiStream(geminiResponse *gemini.GeminiChatResponse, dataChan chan string) {
	// 获取响应中应该使用的模型名称
	responseModel := h.Request.Model
	if h.Context != nil {
		responseModel = base.GetResponseModelNameFromContext(h.Context, h.Request.Model)
	}

	streamResponse := types.ChatCompletionStreamResponse{
		ID:      geminiResponse.ResponseId,
		Object:  "chat.completion.chunk",
		Created: utils.GetTimestamp(),
		Model:   responseModel,
	}

	choices := make([]types.ChatCompletionStreamChoice, 0, len(geminiResponse.Candidates))

	isStop := false
	for _, candidate := range geminiResponse.Candidates {
		if candidate.FinishReason != nil && *candidate.FinishReason == "STOP" {
			isStop = true
			candidate.FinishReason = nil
		}
		choices = append(choices, candidate.ToOpenAIStreamChoice(h.Request))
	}

	if len(choices) > 0 && (choices[0].Delta.ToolCalls != nil || choices[0].Delta.FunctionCall != nil) {
		choices := choices[0].ConvertOpenaiStream()
		for _, choice := range choices {
			chatCompletionCopy := streamResponse
			chatCompletionCopy.Choices = []types.ChatCompletionStreamChoice{choice}
			responseBody, _ := json.Marshal(chatCompletionCopy)
			dataChan <- string(responseBody)
		}
	} else {
		streamResponse.Choices = choices
		responseBody, _ := json.Marshal(streamResponse)
		dataChan <- string(responseBody)
	}

	if isStop {
		streamResponse.Choices = []types.ChatCompletionStreamChoice{
			{
				FinishReason: types.FinishReasonStop,
				Delta: types.ChatCompletionStreamChoiceDelta{
					Role: types.ChatMessageRoleAssistant,
				},
			},
		}
		responseBody, _ := json.Marshal(streamResponse)
		dataChan <- string(responseBody)
	}
}

// Antigravity 默认的 stopSequences
var defaultAntigravityStopSequences = []string{
	"<|user|>",
	"<|bot|>",
	"<|context_request|>",
	"<|endoftext|>",
	"<|end_of_turn|>",
}

// applyAntigravityGenerationConfigDefaults 应用 Antigravity 特有的 generationConfig 默认值
func applyAntigravityGenerationConfigDefaults(requestMap map[string]interface{}) {
	var genConfig map[string]interface{}
	if gc, ok := requestMap["generationConfig"].(map[string]interface{}); ok {
		genConfig = gc
	} else {
		genConfig = make(map[string]interface{})
		requestMap["generationConfig"] = genConfig
	}

	if _, exists := genConfig["topP"]; !exists {
		genConfig["topP"] = float64(1)
	}
	if _, exists := genConfig["topK"]; !exists {
		genConfig["topK"] = float64(40)
	}
	if _, exists := genConfig["candidateCount"]; !exists {
		genConfig["candidateCount"] = 1
	}

	// 合并 stopSequences
	existingStops := []string{}
	if stops, ok := genConfig["stopSequences"].([]interface{}); ok {
		for _, s := range stops {
			if str, ok := s.(string); ok {
				existingStops = append(existingStops, str)
			}
		}
	} else if stops, ok := genConfig["stopSequences"].([]string); ok {
		existingStops = stops
	}
	allStops := make([]string, 0, len(defaultAntigravityStopSequences)+len(existingStops))
	allStops = append(allStops, defaultAntigravityStopSequences...)
	allStops = append(allStops, existingStops...)
	genConfig["stopSequences"] = allStops

	// 设置默认 temperature
	if _, exists := genConfig["temperature"]; !exists {
		genConfig["temperature"] = 0.4
	}
}

// convertToolsToAntigravityFormat 将 tools 结构转换为 Antigravity 格式
// 每个 function declaration 独立包装成一个 functionDeclarations 数组
func convertToolsToAntigravityFormat(requestMap map[string]interface{}) {
	tools, ok := requestMap["tools"].([]interface{})
	if !ok || len(tools) == 0 {
		return
	}

	var allFunctionDecls []interface{}
	var nonFunctionTools []interface{}

	for _, tool := range tools {
		toolMap, ok := tool.(map[string]interface{})
		if !ok {
			continue
		}

		// 收集非 function 类型的 tools（codeExecution, googleSearch, urlContext）
		if _, exists := toolMap["codeExecution"]; exists {
			nonFunctionTools = append(nonFunctionTools, tool)
			continue
		}
		if _, exists := toolMap["googleSearch"]; exists {
			nonFunctionTools = append(nonFunctionTools, tool)
			continue
		}
		if _, exists := toolMap["urlContext"]; exists {
			nonFunctionTools = append(nonFunctionTools, tool)
			continue
		}

		// 提取 functionDeclarations
		if funcDecls, ok := toolMap["functionDeclarations"].([]interface{}); ok {
			allFunctionDecls = append(allFunctionDecls, funcDecls...)
		}
	}

	// 如果没有 functionDeclarations 也没有 nonFunctionTools，无需处理
	if len(allFunctionDecls) == 0 && len(nonFunctionTools) == 0 {
		return
	}

	// 重新构建 tools：每个 function declaration 独立包装
	newTools := make([]interface{}, 0, len(allFunctionDecls)+len(nonFunctionTools))
	for _, funcDecl := range allFunctionDecls {
		newTools = append(newTools, map[string]interface{}{
			"functionDeclarations": []interface{}{funcDecl},
		})
	}
	newTools = append(newTools, nonFunctionTools...)

	// 只有当有有效内容时才更新
	if len(newTools) > 0 {
		requestMap["tools"] = newTools
	}
}

// reorganizeToolMessages 重组消息，确保 functionCall 后紧跟对应的 functionResponse
func reorganizeToolMessages(requestMap map[string]interface{}) {
	contents, ok := requestMap["contents"].([]interface{})
	if !ok || len(contents) == 0 {
		return
	}

	// 收集所有 functionResponse 的 id 映射
	toolResults := make(map[string]interface{})
	for _, content := range contents {
		contentMap, ok := content.(map[string]interface{})
		if !ok {
			continue
		}
		parts, ok := contentMap["parts"].([]interface{})
		if !ok {
			continue
		}
		for _, part := range parts {
			partMap, ok := part.(map[string]interface{})
			if !ok {
				continue
			}
			if funcResp, exists := partMap["functionResponse"]; exists {
				if funcRespMap, ok := funcResp.(map[string]interface{}); ok {
					if id, ok := funcRespMap["id"].(string); ok && id != "" {
						toolResults[id] = part
					}
				}
			}
		}
	}

	if len(toolResults) == 0 {
		return
	}

	// 将消息平铺
	type flatMsg struct {
		role string
		part interface{}
	}
	var flattened []flatMsg

	for _, content := range contents {
		contentMap, ok := content.(map[string]interface{})
		if !ok {
			continue
		}
		role, _ := contentMap["role"].(string)
		if role == "" {
			role = "user"
		}
		parts, ok := contentMap["parts"].([]interface{})
		if !ok {
			continue
		}
		for _, part := range parts {
			flattened = append(flattened, flatMsg{role: role, part: part})
		}
	}

	// 重新组织消息
	var newContents []interface{}

	for i := 0; i < len(flattened); i++ {
		msg := flattened[i]
		partMap, ok := msg.part.(map[string]interface{})
		if !ok {
			newContents = append(newContents, map[string]interface{}{
				"role":  msg.role,
				"parts": []interface{}{msg.part},
			})
			continue
		}

		// 跳过单独的 functionResponse
		if _, exists := partMap["functionResponse"]; exists {
			continue
		}

		// 遇到 functionCall，在其后插入对应的 functionResponse
		if funcCall, exists := partMap["functionCall"]; exists {
			newContents = append(newContents, map[string]interface{}{
				"role":  "model",
				"parts": []interface{}{msg.part},
			})

			if funcCallMap, ok := funcCall.(map[string]interface{}); ok {
				if id, ok := funcCallMap["id"].(string); ok && id != "" {
					if toolResult, exists := toolResults[id]; exists {
						newContents = append(newContents, map[string]interface{}{
							"role":  "user",
							"parts": []interface{}{toolResult},
						})
					}
				}
			}
			continue
		}

		// 其他消息正常添加
		newContents = append(newContents, map[string]interface{}{
			"role":  msg.role,
			"parts": []interface{}{msg.part},
		})
	}

	requestMap["contents"] = newContents
}

// applyToolConfig 当有 functionDeclarations 时添加 toolConfig
func applyToolConfig(requestMap map[string]interface{}) {
	tools, ok := requestMap["tools"].([]interface{})
	if !ok || len(tools) == 0 {
		return
	}

	for _, tool := range tools {
		if toolMap, ok := tool.(map[string]interface{}); ok {
			if _, exists := toolMap["functionDeclarations"]; exists {
				requestMap["toolConfig"] = map[string]interface{}{
					"functionCallingConfig": map[string]interface{}{
						"mode": "VALIDATED",
					},
				}
				return
			}
		}
	}
}
