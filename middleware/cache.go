package middleware

import (
	"github.com/gin-gonic/gin"
)

// NoCache 强制不缓存
func NoCache() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate, private")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		// 不在这里设置 Vary 头部，让 SessionSecurity 中间件统一管理
		c.Next()
	}
}
