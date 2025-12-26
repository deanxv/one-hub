package controller

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"one-api/common"
	"one-api/common/cache"
	"one-api/common/logger"
	"one-api/providers/claudecode"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// ClaudeCode OAuth 状态缓存前缀
	ClaudeCodeOAuthStateCachePrefix = "claudecode_oauth_state:"
	// ClaudeCode OAuth 状态缓存时长（10分钟）
	ClaudeCodeOAuthStateCacheDuration = 10 * time.Minute
)

// ClaudeCode OAuth 配置常量
const (
	ClaudeCodeAuthorizeURL = "https://claude.ai/oauth/authorize"
	ClaudeCodeTokenURL     = "https://console.anthropic.com/v1/oauth/token"
	ClaudeCodeClientID     = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"
	ClaudeCodeRedirectURI  = "https://console.anthropic.com/oauth/code/callback"
	// 使用与 provider 一致的 scopes，确保权限足够
	ClaudeCodeScopes = "user:inference user:profile"
)

// ClaudeCodeOAuthStateData OAuth 状态数据
type ClaudeCodeOAuthStateData struct {
	ChannelID    int    `json:"channel_id"`
	CodeVerifier string `json:"code_verifier"`
	Proxy        string `json:"proxy"` // 代理配置（JSON 字符串）
	CreatedAt    int64  `json:"created_at"`
}

// StartClaudeCodeOAuthRequest 开始 OAuth 认证请求
type StartClaudeCodeOAuthRequest struct {
	ChannelID int    `json:"channel_id"` // 可选，新建时为 0
	Proxy     string `json:"proxy"`      // 可选，代理配置（JSON 字符串）
}

// generateCodeVerifier 生成随机的 code verifier（PKCE）
func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}

// generateCodeChallenge 生成 code challenge（PKCE）
func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
}

// StartClaudeCodeOAuth 开始 ClaudeCode OAuth 认证流程
// POST /api/claudecode/oauth/start
func StartClaudeCodeOAuth(c *gin.Context) {
	var req StartClaudeCodeOAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIRespondWithError(c, http.StatusOK, err)
		return
	}

	// 生成随机 state
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		common.APIRespondWithError(c, http.StatusOK, fmt.Errorf("failed to generate state: %w", err))
		return
	}
	state := base64.URLEncoding.EncodeToString(stateBytes)

	// 生成 PKCE code verifier
	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		common.APIRespondWithError(c, http.StatusOK, fmt.Errorf("failed to generate code verifier: %w", err))
		return
	}

	// 生成 code challenge
	codeChallenge := generateCodeChallenge(codeVerifier)

	// 保存 state 到缓存（包含代理配置）
	stateData := ClaudeCodeOAuthStateData{
		ChannelID:    req.ChannelID,
		CodeVerifier: codeVerifier,
		Proxy:        req.Proxy, // 保存代理配置，用于后续 token 交换
		CreatedAt:    time.Now().Unix(),
	}
	cacheKey := ClaudeCodeOAuthStateCachePrefix + state
	cache.SetCache(cacheKey, stateData, ClaudeCodeOAuthStateCacheDuration)

	// 构建 OAuth 授权 URL
	params := url.Values{}
	params.Set("code", "true")
	params.Set("client_id", ClaudeCodeClientID)
	params.Set("response_type", "code")
	params.Set("redirect_uri", ClaudeCodeRedirectURI)
	params.Set("scope", ClaudeCodeScopes)
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")
	params.Set("state", state)

	authURL := fmt.Sprintf("%s?%s", ClaudeCodeAuthorizeURL, params.Encode())

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"auth_url":   authURL,
			"state":      state,
			"session_id": state, // 使用 state 作为 session_id
			"instructions": []string{
				"1. 点击授权链接，在新窗口中登录 Claude 账户",
				"2. 同意应用权限",
				"3. 授权成功后，复制浏览器地址栏中的完整 URL",
				"4. 将完整的回调 URL 粘贴到下方输入框中",
			},
		},
	})
}

// ExchangeClaudeCodeRequest 交换授权码请求
type ExchangeClaudeCodeRequest struct {
	SessionID         string `json:"session_id"`         // session_id (即 state)
	AuthorizationCode string `json:"authorization_code"` // 授权码或完整的回调 URL
	CallbackURL       string `json:"callback_url"`       // 完整的回调 URL（可选）
}

// ClaudeCodeOAuthCallback 处理用户提交的授权码
// POST /api/claudecode/oauth/exchange-code
func ClaudeCodeOAuthCallback(c *gin.Context) {
	var req ExchangeClaudeCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIRespondWithError(c, http.StatusOK, err)
		return
	}

	if req.SessionID == "" || (req.AuthorizationCode == "" && req.CallbackURL == "") {
		common.APIRespondWithError(c, http.StatusOK, fmt.Errorf("session_id and authorization_code (or callback_url) are required"))
		return
	}

	state := req.SessionID

	// 从缓存中获取 state 数据
	cacheKey := ClaudeCodeOAuthStateCachePrefix + state
	stateData, err := cache.GetCache[ClaudeCodeOAuthStateData](cacheKey)
	if err != nil {
		common.APIRespondWithError(c, http.StatusOK, fmt.Errorf("invalid or expired OAuth session"))
		return
	}

	// 删除已使用的 state
	cache.DeleteCache(cacheKey)

	// 解析授权码（可能是完整的 URL 或直接的 code）
	inputValue := req.CallbackURL
	if inputValue == "" {
		inputValue = req.AuthorizationCode
	}

	code, err := parseCallbackURL(inputValue)
	if err != nil {
		common.APIRespondWithError(c, http.StatusOK, fmt.Errorf("failed to parse authorization code: %w", err))
		return
	}

	// 使用 code 交换 token（使用会话中保存的代理配置）
	tokenResp, err := exchangeClaudeCodeForToken(code, stateData.CodeVerifier, state, stateData.Proxy)
	if err != nil {
		logger.SysError(fmt.Sprintf("Failed to exchange code for token: %s", err.Error()))
		common.APIRespondWithError(c, http.StatusOK, fmt.Errorf("failed to exchange code for token: %w", err))
		return
	}

	// 构建凭证对象
	credentials := &claudecode.OAuth2Credentials{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ClientID:     ClaudeCodeClientID,
		TokenType:    tokenResp.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	// 解析 scopes
	if tokenResp.Scope != "" {
		credentials.Scopes = strings.Split(tokenResp.Scope, " ")
	}

	// 序列化凭证
	credentialsJSON, err := credentials.ToJSON()
	if err != nil {
		logger.SysError(fmt.Sprintf("Failed to serialize credentials: %s", err.Error()))
		common.APIRespondWithError(c, http.StatusOK, fmt.Errorf("failed to serialize credentials: %w", err))
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "授权成功",
		"data": gin.H{
			"credentials": credentialsJSON,
		},
	})
}

// parseCallbackURL 解析回调 URL 或授权码
func parseCallbackURL(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("empty input")
	}

	trimmedInput := strings.TrimSpace(input)

	// 情况1: 尝试作为完整URL解析
	if strings.HasPrefix(trimmedInput, "http://") || strings.HasPrefix(trimmedInput, "https://") {
		parsedURL, err := url.Parse(trimmedInput)
		if err != nil {
			return "", fmt.Errorf("invalid URL format: %w", err)
		}

		code := parsedURL.Query().Get("code")
		if code == "" {
			return "", fmt.Errorf("code parameter not found in callback URL")
		}

		return code, nil
	}

	// 情况2: 直接的授权码（可能包含URL fragments）
	cleanedCode := strings.Split(strings.Split(trimmedInput, "#")[0], "&")[0]

	// 验证授权码格式
	if len(cleanedCode) < 10 {
		return "", fmt.Errorf("authorization code too short")
	}

	return cleanedCode, nil
}

// exchangeClaudeCodeForToken 使用授权码交换访问令牌（支持代理）
func exchangeClaudeCodeForToken(code, codeVerifier, state, proxyURL string) (*claudecode.TokenRefreshResponse, error) {
	// 准备请求数据（使用 JSON 格式）
	requestBody := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     ClaudeCodeClientID,
		"code":          code,
		"redirect_uri":  ClaudeCodeRedirectURI,
		"code_verifier": codeVerifier,
		"state":         state,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", ClaudeCodeTokenURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "claude-cli/1.0.56 (external, cli)")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://claude.ai/")
	req.Header.Set("Origin", "https://claude.ai")

	// 创建 HTTP 客户端
	client := &http.Client{Timeout: 30 * time.Second}

	// 如果有代理配置，设置代理
	if proxyURL != "" {
		proxyURLParsed, err := url.Parse(proxyURL)
		if err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURLParsed),
			}
			logger.SysLog(fmt.Sprintf("Using proxy for ClaudeCode token exchange: %s", proxyURL))
		} else {
			logger.SysError(fmt.Sprintf("Failed to parse proxy URL: %s", err.Error()))
		}
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 解析响应
	var tokenResp claudecode.TokenRefreshResponse
	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &tokenResp, nil
}
