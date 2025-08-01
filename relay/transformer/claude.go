package transformer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/providers/claude"
	"strings"
	"time"
)

// ToolCallState 用于管理流式工具调用的状态
type ToolCallState struct {
	Id                string
	Name              string
	Arguments         string
	ContentBlockIndex int
}

// ClaudeTransformer Claude 格式转换器
type ClaudeTransformer struct {
	name string
}

// NewClaudeTransformer 创建 Claude 转换器
func NewClaudeTransformer() *ClaudeTransformer {
	return &ClaudeTransformer{
		name: "claude",
	}
}

// GetName 获取转换器名称
func (t *ClaudeTransformer) GetName() string {
	return t.name
}

// TransformRequestOut 将 Claude 请求转换为统一格式
func (t *ClaudeTransformer) TransformRequestOut(request interface{}) (*UnifiedChatRequest, error) {
	claudeReq, ok := request.(*claude.ClaudeRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for Claude transformer")
	}

	unified := &UnifiedChatRequest{
		Model:     claudeReq.Model,
		MaxTokens: claudeReq.MaxTokens,
		Stream:    claudeReq.Stream,
		System:    claudeReq.System,
	}

	if claudeReq.Temperature != nil {
		unified.Temperature = claudeReq.Temperature
	}

	// 转换消息
	for _, msg := range claudeReq.Messages {
		unifiedMsg := UnifiedMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
		unified.Messages = append(unified.Messages, unifiedMsg)
	}

	// 转换工具
	if len(claudeReq.Tools) > 0 {
		for _, tool := range claudeReq.Tools {
			// 处理 InputSchema 的类型转换
			var parameters map[string]interface{}
			if tool.InputSchema != nil {
				if paramMap, ok := tool.InputSchema.(map[string]interface{}); ok {
					parameters = paramMap
				} else {
					// 如果不是 map[string]interface{}，尝试通过 JSON 转换
					if jsonBytes, err := json.Marshal(tool.InputSchema); err == nil {
						json.Unmarshal(jsonBytes, &parameters)
					}
				}
			}
			if parameters == nil {
				parameters = make(map[string]interface{})
			}

			unifiedTool := UnifiedTool{
				Type: "function",
				Function: UnifiedToolFunction{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  parameters,
				},
			}
			unified.Tools = append(unified.Tools, unifiedTool)
		}
	}

	if claudeReq.ToolChoice != nil {
		unified.ToolChoice = claudeReq.ToolChoice
	}

	return unified, nil
}

// TransformRequestIn 将统一格式转换为 Claude 请求格式（不需要实现）
func (t *ClaudeTransformer) TransformRequestIn(request *UnifiedChatRequest) (interface{}, error) {
	return nil, fmt.Errorf("Claude transformer does not support TransformRequestIn")
}

// TransformResponseOut 将 Claude 响应转换为统一格式（不需要实现）
func (t *ClaudeTransformer) TransformResponseOut(response *http.Response) (*UnifiedChatResponse, error) {
	return nil, fmt.Errorf("Claude transformer does not support TransformResponseOut")
}

// TransformResponseIn 将统一格式转换为 Claude 响应格式
func (t *ClaudeTransformer) TransformResponseIn(response *UnifiedChatResponse) (interface{}, error) {

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices in unified response")
	}

	choice := response.Choices[0]
	var content []claude.ResContent

	// 处理消息内容
	if choice.Message != nil && choice.Message.Content != nil {
		if contentStr, ok := choice.Message.Content.(string); ok && contentStr != "" {
			content = append(content, claude.ResContent{
				Type: "text",
				Text: contentStr,
			})
		}
	}

	// 处理工具调用
	if choice.Message != nil && len(choice.Message.ToolCalls) > 0 {
		for _, toolCall := range choice.Message.ToolCalls {
			var input interface{}
			if toolCall.Function.Arguments != "" {
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &input); err != nil {
					input = map[string]interface{}{
						"arguments": toolCall.Function.Arguments,
					}
				}
			} else {
				input = map[string]interface{}{}
			}

			content = append(content, claude.ResContent{
				Type:  "tool_use",
				Id:    toolCall.Id,
				Name:  toolCall.Function.Name,
				Input: input,
			})
		}
	}

	// 转换停止原因
	stopReason := "end_turn"
	if choice.FinishReason != "" {
		switch choice.FinishReason {
		case "stop":
			stopReason = "end_turn"
		case "length":
			stopReason = "max_tokens"
		case "tool_calls":
			stopReason = "tool_use"
		case "content_filter":
			stopReason = "stop_sequence"
		}
	}

	claudeResponse := &claude.ClaudeResponse{
		Id:         response.Id,
		Type:       "message",
		Role:       "assistant",
		Content:    content,
		Model:      response.Model,
		StopReason: stopReason,
	}

	// 处理 usage 信息
	if response.Usage != nil {
		finalOutputTokens := response.Usage.CompletionTokens

		// 如果 CompletionTokens 为 0，尝试从内容计算
		if finalOutputTokens == 0 {
			var textContent strings.Builder
			var toolCallTokens int

			// 计算文本内容 tokens
			for _, c := range content {
				if c.Type == "text" && c.Text != "" {
					textContent.WriteString(c.Text)
				} else if c.Type == "tool_use" {
					// 计算工具调用 tokens
					toolCallText := fmt.Sprintf("tool_use:%s:%v", c.Name, c.Input)
					toolCallTokens += common.CountTokenText(toolCallText, response.Model)
				}
			}

			// 累加工具调用和文本内容的 tokens
			finalOutputTokens = toolCallTokens
			if textContent.Len() > 0 {
				textTokens := common.CountTokenText(textContent.String(), response.Model)
				finalOutputTokens += textTokens
			}
		}

		claudeResponse.Usage = claude.Usage{
			InputTokens:  response.Usage.PromptTokens,
			OutputTokens: finalOutputTokens,
		}
	}

	return claudeResponse, nil
}

// TransformStreamResponseOut 将 Claude 流式响应转换为统一格式流（不需要实现）
func (t *ClaudeTransformer) TransformStreamResponseOut(response *http.Response) (*http.Response, error) {
	return nil, fmt.Errorf("Claude transformer does not support TransformStreamResponseOut")
}

// TransformStreamResponseIn converts unified format stream to Claude stream response format
func (t *ClaudeTransformer) TransformStreamResponseIn(response *http.Response) (*http.Response, error) {
	if response.Body == nil {
		return response, nil
	}

	// create pipe
	pr, pw := io.Pipe()

	go func() {
		defer func() {
			pw.Close()
		}()
		defer response.Body.Close()

		scanner := bufio.NewScanner(response.Body)
		messageId := fmt.Sprintf("msg_%d", time.Now().UnixNano())
		hasStarted := false
		hasTextContentStarted := false
		contentIndex := 0

		// 工具调用状态管理（按照 demo 的方式）
		toolCalls := make(map[int]*ToolCallState)
		toolCallIndexToContentBlockIndex := make(map[int]int)
		hasFinished := false // 流是否已结束

		// 累积工具调用的 token 数（用于当上游不提供 usage 时的计算）
		toolCallTokens := 0
		var textBuilder strings.Builder

		// 发送 message_start 事件
		if !hasStarted {
			hasStarted = true
			messageStart := map[string]interface{}{
				"type": "message_start",
				"message": map[string]interface{}{
					"id":            messageId,
					"type":          "message",
					"role":          "assistant",
					"content":       []interface{}{},
					"model":         "unknown",
					"stop_reason":   nil,
					"stop_sequence": nil,
					"usage": map[string]interface{}{
						"input_tokens":  0,
						"output_tokens": 0,
					},
				},
			}
			t.writeSSEEvent(pw, "message_start", messageStart)
		}

		for scanner.Scan() {
			line := scanner.Text()

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			if data == "[DONE]" {
				continue
			}

			var chunk UnifiedChatResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) == 0 {
				continue
			}

			choice := chunk.Choices[0]

			// 处理文本内容
			if choice.Delta != nil && choice.Delta.Content != nil {
				if contentStr, ok := choice.Delta.Content.(string); ok && contentStr != "" {
					// 累积文本内容用于 token 计算
					textBuilder.WriteString(contentStr)

					if !hasTextContentStarted {
						hasTextContentStarted = true
						contentBlockStart := map[string]interface{}{
							"type":  "content_block_start",
							"index": contentIndex,
							"content_block": map[string]interface{}{
								"type": "text",
								"text": "",
							},
						}
						t.writeSSEEvent(pw, "content_block_start", contentBlockStart)
					}

					contentBlockDelta := map[string]interface{}{
						"type":  "content_block_delta",
						"index": contentIndex,
						"delta": map[string]interface{}{
							"type": "text_delta",
							"text": contentStr,
						},
					}
					t.writeSSEEvent(pw, "content_block_delta", contentBlockDelta)
				}
			}

			// handle tool calls (following demo logic)
			if choice.Delta != nil && len(choice.Delta.ToolCalls) > 0 {
				processedInThisChunk := make(map[int]bool)

				// 我们需要从原始数据中获取 index，而不是从 UnifiedToolCall 中获取
				// 因为 UnifiedToolCall 没有 Index 字段
				// 让我们从原始的 JSON 数据中解析
				choiceBytes, _ := json.Marshal(choice)
				var rawChoice map[string]interface{}
				json.Unmarshal(choiceBytes, &rawChoice)

				if delta, exists := rawChoice["delta"].(map[string]interface{}); exists {
					if toolCallsData, exists := delta["tool_calls"].([]interface{}); exists {
						for i, tc := range toolCallsData {
							tcMap, ok := tc.(map[string]interface{})
							if !ok {
								continue
							}

							// 获取 index
							toolCallIndex := i // 默认使用数组索引
							if idx, exists := tcMap["index"]; exists && idx != nil {
								if idxFloat, ok := idx.(float64); ok {
									toolCallIndex = int(idxFloat)
								}
							}

							// 获取其他字段
							var toolCallId, toolCallName, toolCallArgs string
							if id, exists := tcMap["id"]; exists && id != nil {
								toolCallId = id.(string)
							}
							if function, exists := tcMap["function"].(map[string]interface{}); exists {
								if name, exists := function["name"]; exists && name != nil {
									toolCallName = name.(string)
								}
								if args, exists := function["arguments"]; exists && args != nil {
									toolCallArgs = args.(string)
								}
							}

							// avoid duplicate processing of same index in same chunk
							if processedInThisChunk[toolCallIndex] {
								continue
							}
							processedInThisChunk[toolCallIndex] = true

							// check if this is a new tool call index
							_, isKnownIndex := toolCallIndexToContentBlockIndex[toolCallIndex]

							if !isKnownIndex {
								// first time encountering this index, create new content block
								if hasTextContentStarted {
									contentBlockStop := map[string]interface{}{
										"type":  "content_block_stop",
										"index": contentIndex,
									}
									t.writeSSEEvent(pw, "content_block_stop", contentBlockStop)
									contentIndex++
									hasTextContentStarted = false
								}

								// 计算新的 content block index
								newContentBlockIndex := len(toolCallIndexToContentBlockIndex)
								if hasTextContentStarted {
									newContentBlockIndex = len(toolCallIndexToContentBlockIndex) + 1
								}

								toolCallIndexToContentBlockIndex[toolCallIndex] = newContentBlockIndex

								// 生成工具调用 ID 和名称
								if toolCallId == "" {
									toolCallId = fmt.Sprintf("call_%d_%d", time.Now().UnixNano(), toolCallIndex)
								}

								if toolCallName == "" {
									toolCallName = fmt.Sprintf("tool_%d", toolCallIndex)
								}

								// send content_block_start
								contentBlockStart := map[string]interface{}{
									"type":  "content_block_start",
									"index": newContentBlockIndex,
									"content_block": map[string]interface{}{
										"type":  "tool_use",
										"id":    toolCallId,
										"name":  toolCallName,
										"input": map[string]interface{}{},
									},
								}
								t.writeSSEEvent(pw, "content_block_start", contentBlockStart)

								// 创建工具调用状态
								toolCalls[toolCallIndex] = &ToolCallState{
									Id:                toolCallId,
									Name:              toolCallName,
									Arguments:         "",
									ContentBlockIndex: newContentBlockIndex,
								}

								contentIndex = newContentBlockIndex
							} else if toolCallId != "" && toolCallName != "" {
								// 更新已存在的工具调用的 ID 和名称（如果之前是临时的）
								existingToolCall := toolCalls[toolCallIndex]
								if existingToolCall != nil {
									wasTemporary := strings.HasPrefix(existingToolCall.Id, "call_") && strings.HasPrefix(existingToolCall.Name, "tool_")
									if wasTemporary {
										existingToolCall.Id = toolCallId
										existingToolCall.Name = toolCallName
									}
								}
							}

							// handle argument increments
							if toolCallArgs != "" && toolCallArgs != "{}" && !hasFinished {
								blockIndex, exists := toolCallIndexToContentBlockIndex[toolCallIndex]
								if !exists {
									continue
								}

								currentToolCall := toolCalls[toolCallIndex]
								if currentToolCall != nil {
									currentToolCall.Arguments += toolCallArgs
								}

								// send content_block_delta
								contentBlockDelta := map[string]interface{}{
									"type":  "content_block_delta",
									"index": blockIndex,
									"delta": map[string]interface{}{
										"type":         "input_json_delta",
										"partial_json": toolCallArgs,
									},
								}
								t.writeSSEEvent(pw, "content_block_delta", contentBlockDelta)
							}
						}
					}
				}
			}

			// 处理结束
			if choice.FinishReason != "" {

				if hasTextContentStarted {
					contentBlockStop := map[string]interface{}{
						"type":  "content_block_stop",
						"index": contentIndex,
					}
					t.writeSSEEvent(pw, "content_block_stop", contentBlockStop)
				}

				// 为所有工具调用发送 content_block_stop 并计算 tokens
				for _, toolCallState := range toolCalls {
					contentBlockStop := map[string]interface{}{
						"type":  "content_block_stop",
						"index": toolCallState.ContentBlockIndex,
					}
					t.writeSSEEvent(pw, "content_block_stop", contentBlockStop)

					// 计算工具调用的 token 数（在工具调用完成时）
					if toolCallState.Name != "" && toolCallState.Arguments != "" {
						toolCallText := fmt.Sprintf("tool_use:%s:%s", toolCallState.Name, toolCallState.Arguments)
						toolCallTokens += common.CountTokenText(toolCallText, "gpt-3.5-turbo")
					}
				}

				// 转换停止原因
				stopReason := "end_turn"
				switch choice.FinishReason {
				case "stop":
					stopReason = "end_turn"
				case "length":
					stopReason = "max_tokens"
				case "tool_calls":
					stopReason = "tool_use"
				case "content_filter":
					stopReason = "stop_sequence"
				}

				messageDelta := map[string]interface{}{
					"type": "message_delta",
					"delta": map[string]interface{}{
						"stop_reason":   stopReason,
						"stop_sequence": nil,
					},
				}

				// 始终包含usage信息，即使为0
				if chunk.Usage != nil {
					messageDelta["usage"] = map[string]interface{}{
						"input_tokens":  chunk.Usage.PromptTokens,
						"output_tokens": chunk.Usage.CompletionTokens,
					}
				} else {
					// 如果没有usage信息，尝试使用累积的 tokens
					estimatedOutputTokens := toolCallTokens

					// 累加文本内容的 tokens
					if textBuilder.Len() > 0 {
						textTokens := common.CountTokenText(textBuilder.String(), "gpt-3.5-turbo")
						estimatedOutputTokens += textTokens
					}

					messageDelta["usage"] = map[string]interface{}{
						"input_tokens":  0,
						"output_tokens": estimatedOutputTokens,
					}
				}

				t.writeSSEEvent(pw, "message_delta", messageDelta)

				messageStop := map[string]interface{}{
					"type": "message_stop",
				}
				t.writeSSEEvent(pw, "message_stop", messageStop)
				break
			}
		}

		if err := scanner.Err(); err != nil {
			// log scan error if needed
		}
	}()

	// 创建新的响应
	newResponse := &http.Response{
		Status:        response.Status,
		StatusCode:    response.StatusCode,
		Proto:         response.Proto,
		ProtoMajor:    response.ProtoMajor,
		ProtoMinor:    response.ProtoMinor,
		Header:        make(http.Header),
		Body:          pr,
		ContentLength: -1,
	}

	// 复制头部
	for k, v := range response.Header {
		newResponse.Header[k] = v
	}

	// 设置流式响应头部
	newResponse.Header.Set("Content-Type", "text/event-stream")
	newResponse.Header.Set("Cache-Control", "no-cache")
	newResponse.Header.Set("Connection", "keep-alive")

	return newResponse, nil
}

// writeSSEEvent writes SSE event
func (t *ClaudeTransformer) writeSSEEvent(w io.Writer, event string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	eventData := fmt.Sprintf("event: %s\ndata: %s\n\n", event, string(jsonData))
	fmt.Fprintf(w, "%s", eventData)
}
