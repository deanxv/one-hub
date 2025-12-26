package claudecode

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/requester"
	"one-api/providers/claude"
	"one-api/types"
	"strings"

	"github.com/google/uuid"
)

// CreateChatCompletion 创建聊天完成
func (p *ClaudeCodeProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	request.OneOtherArg = p.GetOtherArg()
	claudeRequest, errWithCode := claude.ConvertFromChatOpenai(request)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 应用 ClaudeCode 兼容性处理
	p.applyClaudeCodeCompatibility(claudeRequest)

	req, errWithCode := p.getChatRequest(claudeRequest)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	claudeResponse := &claude.ClaudeResponse{}
	_, errWithCode = p.Requester.SendRequest(req, claudeResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return claude.ConvertToChatOpenai(&p.ClaudeProvider, claudeResponse, request)
}

// CreateChatCompletionStream 创建流式聊天完成
func (p *ClaudeCodeProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
	request.OneOtherArg = p.GetOtherArg()
	claudeRequest, errWithCode := claude.ConvertFromChatOpenai(request)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 应用 ClaudeCode 兼容性处理
	p.applyClaudeCodeCompatibility(claudeRequest)

	req, errWithCode := p.getChatRequest(claudeRequest)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	resp, errWithCode := p.Requester.SendRequestRaw(req)
	if errWithCode != nil {
		return nil, errWithCode
	}

	chatHandler := &claude.ClaudeStreamHandler{
		Usage:   p.Usage,
		Request: request,
		Prefix:  `data: {"type"`,
		Context: p.Context,
	}

	return requester.RequestStream(p.Requester, resp, chatHandler.HandlerStream)
}

// getChatRequest 获取聊天请求
func (p *ClaudeCodeProvider) getChatRequest(claudeRequest *claude.ClaudeRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(config.RelayModeChatCompletions)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url)
	if fullRequestURL == "" {
		return nil, common.ErrorWrapperLocal(nil, "invalid_claudecode_config", http.StatusInternalServerError)
	}

	// 获取请求头
	headers := p.GetRequestHeaders()

	// 检查 token 是否获取成功
	if _, hasAuth := headers["Authorization"]; !hasAuth {
		// Token 获取失败，返回详细错误信息
		token, err := p.GetToken()
		if err != nil {
			return nil, p.handleTokenError(err)
		}
		// 如果 GetToken 成功但 headers 中没有 Authorization，手动添加
		headers["Authorization"] = "Bearer " + token
	}

	// 应用 ClaudeCode 默认请求头
	p.applyDefaultHeaders(headers)

	if claudeRequest.Stream {
		headers["Accept"] = "text/event-stream"
	}

	// 使用BaseProvider的统一方法创建请求，支持额外参数处理
	req, errWithCode := p.NewRequestWithCustomParams(http.MethodPost, fullRequestURL, claudeRequest, headers, claudeRequest.Model)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return req, nil
}

// applyDefaultHeaders 应用 ClaudeCode 默认请求头
func (p *ClaudeCodeProvider) applyDefaultHeaders(headers map[string]string) {
	// 如果没有 anthropic-beta，设置默认值
	if _, exists := headers["anthropic-beta"]; !exists {
		headers["anthropic-beta"] = "claude-code-20250219,oauth-2025-04-20,interleaved-thinking-2025-05-14,fine-grained-tool-streaming-2025-05-14"
	}

	// 如果没有 user-agent，设置默认值
	if _, exists := headers["user-agent"]; !exists {
		headers["user-agent"] = "claude-cli/1.0.81 (external, cli)"
	}

	// 添加 ClaudeCode 必需的 x-stainless-* 头部
	if _, exists := headers["x-stainless-retry-count"]; !exists {
		headers["x-stainless-retry-count"] = "0"
	}
	if _, exists := headers["x-stainless-timeout"]; !exists {
		headers["x-stainless-timeout"] = "60"
	}
	if _, exists := headers["x-stainless-lang"]; !exists {
		headers["x-stainless-lang"] = "js"
	}
	if _, exists := headers["x-stainless-package-version"]; !exists {
		headers["x-stainless-package-version"] = "0.55.1"
	}
	if _, exists := headers["x-stainless-os"]; !exists {
		headers["x-stainless-os"] = "Windows"
	}
	if _, exists := headers["x-stainless-arch"]; !exists {
		headers["x-stainless-arch"] = "x64"
	}
	if _, exists := headers["x-stainless-runtime"]; !exists {
		headers["x-stainless-runtime"] = "node"
	}
	if _, exists := headers["x-stainless-runtime-version"]; !exists {
		headers["x-stainless-runtime-version"] = "v20.19.2"
	}

	// 添加其他必需的头部
	if _, exists := headers["x-app"]; !exists {
		headers["x-app"] = "cli"
	}
	if _, exists := headers["anthropic-dangerous-direct-browser-access"]; !exists {
		headers["anthropic-dangerous-direct-browser-access"] = "true"
	}
	if _, exists := headers["accept-language"]; !exists {
		headers["accept-language"] = "*"
	}
	if _, exists := headers["sec-fetch-mode"]; !exists {
		headers["sec-fetch-mode"] = "cors"
	}
}

// generateClaudeCodeUserId 生成 ClaudeCode 格式的 user_id
// 格式: user_{64位十六进制}_account__session_{uuid}
func generateClaudeCodeUserId() string {
	// 生成一个随机的64位十六进制字符串
	hash := sha256.New()
	sessionUUID := uuid.New().String()
	hash.Write([]byte(sessionUUID))
	userHash := hex.EncodeToString(hash.Sum(nil))

	// 生成session UUID
	sessionID := uuid.New().String()

	// 组合成 ClaudeCode 格式
	return fmt.Sprintf("user_%s_account__session_%s", userHash, sessionID)
}

// extractMetadataFromOriginalRequest 从原始请求体中提取 metadata 字段
func (p *ClaudeCodeProvider) extractMetadataFromOriginalRequest(claudeRequest *claude.ClaudeRequest) {
	if p.Context == nil {
		return
	}

	// 从 gin.Context 中获取原始请求体
	rawBody, exists := p.Context.Get(config.GinRequestBodyKey)
	if !exists {
		return
	}

	bodyBytes, ok := rawBody.([]byte)
	if !ok {
		return
	}

	// 解析原始请求体为 map
	var requestMap map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &requestMap); err != nil {
		return
	}

	// 提取 metadata 字段
	if metadataInterface, exists := requestMap["metadata"]; exists {
		if metadataMap, ok := metadataInterface.(map[string]interface{}); ok {
			// 提取 user_id
			if userID, ok := metadataMap["user_id"].(string); ok && userID != "" {
				if claudeRequest.Metadata == nil {
					claudeRequest.Metadata = &claude.ClaudeMetadata{}
				}
				claudeRequest.Metadata.UserId = userID
			}
		}
	}
}

// applyClaudeCodeCompatibility 应用 ClaudeCode 兼容性处理
// 确保 system 字段中包含必需的 "You are Claude Code, Anthropic's official CLI for Claude." 缓存控制项
// 并添加 metadata.user_id
func (p *ClaudeCodeProvider) applyClaudeCodeCompatibility(claudeRequest *claude.ClaudeRequest) {
	// 首先尝试从原始请求体中提取 metadata
	p.extractMetadataFromOriginalRequest(claudeRequest)
	// 必需的缓存控制项
	requiredCacheItem := claude.MessageContent{
		Type: "text",
		Text: "You are Claude Code, Anthropic's official CLI for Claude.",
		CacheControl: map[string]string{
			"type": "ephemeral",
		},
	}

	// 检查是否已存在该缓存控制项
	hasRequiredItem := false

	// 将 system 转换为统一的 []MessageContent 格式
	var systemContents []claude.MessageContent

	if claudeRequest.System == nil || claudeRequest.System == "" {
		// 情况1: system 为空，直接使用必需项
		systemContents = []claude.MessageContent{requiredCacheItem}
		claudeRequest.System = systemContents
		return
	}

	if systemStr, ok := claudeRequest.System.(string); ok {
		// 情况2: system 是字符串
		if strings.TrimSpace(systemStr) != "" {
			// 保留原有字符串内容
			systemContents = append(systemContents, claude.MessageContent{
				Type: "text",
				Text: systemStr,
			})
		}
	} else if systemArray, ok := claudeRequest.System.([]interface{}); ok {
		// 情况3: system 是 []interface{} 类型
		for _, item := range systemArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemType, ok := itemMap["type"].(string); ok && itemType == "text" {
					if text, ok := itemMap["text"].(string); ok {
						// 检查是否是必需的缓存控制项
						if text == "You are Claude Code, Anthropic's official CLI for Claude." {
							if cacheControl, exists := itemMap["cache_control"].(map[string]interface{}); exists {
								if cacheType, ok := cacheControl["type"].(string); ok && cacheType == "ephemeral" {
									hasRequiredItem = true
								}
							}
						}

						// 保留所有内容
						content := claude.MessageContent{
							Type: "text",
							Text: text,
						}
						if cacheControl, exists := itemMap["cache_control"].(map[string]interface{}); exists {
							cacheControlMap := make(map[string]string)
							for k, v := range cacheControl {
								if strVal, ok := v.(string); ok {
									cacheControlMap[k] = strVal
								}
							}
							content.CacheControl = cacheControlMap
						}
						systemContents = append(systemContents, content)
					}
				}
			}
		}
	} else if systemArray, ok := claudeRequest.System.([]claude.MessageContent); ok {
		// 情况4: system 是 []MessageContent 类型
		for _, item := range systemArray {
			// 检查是否是必需的缓存控制项
			if item.Type == "text" && item.Text == "You are Claude Code, Anthropic's official CLI for Claude." {
				if item.CacheControl != nil {
					if cacheControlMap, ok := item.CacheControl.(map[string]string); ok {
						if cacheType, exists := cacheControlMap["type"]; exists && cacheType == "ephemeral" {
							hasRequiredItem = true
						}
					} else if cacheControlMap, ok := item.CacheControl.(map[string]interface{}); ok {
						if cacheType, exists := cacheControlMap["type"].(string); exists && cacheType == "ephemeral" {
							hasRequiredItem = true
						}
					}
				}
			}
			// 保留所有内容
			systemContents = append(systemContents, item)
		}
	}

	// 如果不存在必需的缓存控制项，添加到开头
	if !hasRequiredItem {
		systemContents = append([]claude.MessageContent{requiredCacheItem}, systemContents...)
	}

	// 更新 system 字段
	claudeRequest.System = systemContents

	// 添加 metadata.user_id（如果不存在）
	if claudeRequest.Metadata == nil {
		claudeRequest.Metadata = &claude.ClaudeMetadata{
			UserId: generateClaudeCodeUserId(),
		}
	} else if claudeRequest.Metadata.UserId == "" {
		claudeRequest.Metadata.UserId = generateClaudeCodeUserId()
	}
}
