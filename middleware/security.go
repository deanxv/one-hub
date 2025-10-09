package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// SecurityHeaders 添加通用的安全头部
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 添加通用的安全头部
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")

		c.Next()
	}
}

// SessionSecurity 为会话相关的接口添加额外的安全头部
func SessionSecurity() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 添加会话安全相关的头部
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// 设置强制防缓存策略
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate, private, max-age=0")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		// 检查是否使用 Cloudflare，设置相应的 Vary 头部
		trustedHeader := viper.GetString("trusted_header")
		if trustedHeader == "CF-Connecting-IP" {
			// 对于 Cloudflare，添加额外的 Vary 头部确保不同用户不会共享缓存
			c.Header("Vary", "Cookie, Authorization, X-Requested-With, User-Agent, CF-Connecting-IP")
		} else {
			c.Header("Vary", "Cookie, Authorization, X-Requested-With, User-Agent")
		}

		c.Next()
	}
}
