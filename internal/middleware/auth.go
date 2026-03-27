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
		// 如果配置中没有设置 key，则直接放行 (为了兼容性)
		if expectedKey == "" {
			c.Next()
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
