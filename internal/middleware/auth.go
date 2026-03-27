package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/pkg/errors"
	"github.com/jianyun8023/calibre-api/pkg/response"
)

// APIKeyAuth 验证请求是否带有正确的 API Key
func APIKeyAuth(expectedKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 鉴权绕过漏洞修复：如果未配置 API Key，直接拒绝服务，强制要求配置
		if expectedKey == "" {
			err := errors.New(errors.CodeUnauthorized, "API Key is not configured on the server", http.StatusUnauthorized)
			response.Error(c, err)
			c.Abort()
			return
		}

		// 检查 Header 中的 X-API-Key 或者 Authorization: Bearer
		authHeader := c.GetHeader("Authorization")
		apiKey := c.GetHeader("X-API-Key")

		token := ""
		if apiKey != "" {
			token = apiKey
		} else if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

		if token == "" || token != expectedKey {
			err := errors.New(errors.CodeUnauthorized, "Invalid or missing API Key", http.StatusUnauthorized)
			response.Error(c, err)
			c.Abort()
			return
		}

		c.Next()
	}
}
