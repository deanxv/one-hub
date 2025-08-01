package transformer

import (
	"net/http"
	"one-api/types"
)

// UnifiedChatRequest 统一的聊天请求格式
type UnifiedChatRequest struct {
	Model       string           `json:"model"`
	Messages    []UnifiedMessage `json:"messages"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Temperature *float64         `json:"temperature,omitempty"`
	Stream      bool             `json:"stream,omitempty"`
	Tools       []UnifiedTool    `json:"tools,omitempty"`
	ToolChoice  interface{}      `json:"tool_choice,omitempty"`
	System      interface{}      `json:"system,omitempty"`
}

// UnifiedMessage 统一的消息格式
type UnifiedMessage struct {
	Role         string                 `json:"role"`
	Content      interface{}            `json:"content"`
	ToolCalls    []UnifiedToolCall      `json:"tool_calls,omitempty"`
	ToolCallId   string                 `json:"tool_call_id,omitempty"`
	CacheControl map[string]interface{} `json:"cache_control,omitempty"`
}

// UnifiedTool 统一的工具格式
type UnifiedTool struct {
	Type     string              `json:"type"`
	Function UnifiedToolFunction `json:"function"`
}

// UnifiedToolFunction 统一的工具函数格式
type UnifiedToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// UnifiedToolCall 统一的工具调用格式
type UnifiedToolCall struct {
	Id       string                  `json:"id"`
	Type     string                  `json:"type"`
	Function UnifiedToolCallFunction `json:"function"`
}

// UnifiedToolCallFunction 统一的工具调用函数格式
type UnifiedToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// UnifiedChatResponse 统一的聊天响应格式
type UnifiedChatResponse struct {
	Id      string          `json:"id"`
	Object  string          `json:"object"`
	Created int64           `json:"created"`
	Model   string          `json:"model"`
	Choices []UnifiedChoice `json:"choices"`
	Usage   *types.Usage    `json:"usage,omitempty"`
}

// UnifiedChoice 统一的选择格式
type UnifiedChoice struct {
	Index        int             `json:"index"`
	Message      *UnifiedMessage `json:"message,omitempty"`
	Delta        *UnifiedMessage `json:"delta,omitempty"`
	FinishReason string          `json:"finish_reason,omitempty"`
}

// Transformer 转换器接口
type Transformer interface {
	// GetName 获取转换器名称
	GetName() string

	// TransformRequestOut 将原始请求转换为统一格式
	TransformRequestOut(request interface{}) (*UnifiedChatRequest, error)

	// TransformRequestIn 将统一格式转换为目标API格式
	TransformRequestIn(request *UnifiedChatRequest) (interface{}, error)

	// TransformResponseOut 将目标API响应转换为统一格式
	TransformResponseOut(response *http.Response) (*UnifiedChatResponse, error)

	// TransformResponseIn 将统一格式转换为原始响应格式
	TransformResponseIn(response *UnifiedChatResponse) (interface{}, error)

	// TransformStreamResponseOut 将目标API流式响应转换为统一格式流
	TransformStreamResponseOut(response *http.Response) (*http.Response, error)

	// TransformStreamResponseIn 将统一格式流转换为原始流式响应格式
	TransformStreamResponseIn(response *http.Response) (*http.Response, error)
}
