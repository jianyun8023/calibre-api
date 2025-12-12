package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	// RequestIDKey context key for request ID
	RequestIDKey = "request_id"
)

// GinLogger Gin 日志中间件
// 记录每个 HTTP 请求的详细信息
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成请求 ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(RequestIDKey, requestID)

		// 记录请求开始时间
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算请求耗时
		latency := time.Since(start)

		// 获取响应信息
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// 构建日志字段
		event := globalLogger.Info()
		if statusCode >= 500 {
			event = globalLogger.Error()
		} else if statusCode >= 400 {
			event = globalLogger.Warn()
		}

		event.
			Str("request_id", requestID).
			Str("client_ip", clientIP).
			Str("method", method).
			Str("path", path).
			Str("query", raw).
			Int("status", statusCode).
			Dur("latency", latency).
			Str("latency_human", latency.String()).
			Int("body_size", c.Writer.Size())

		// 添加错误信息
		if errorMessage != "" {
			event.Str("error", errorMessage)
		}

		// 添加用户信息（如果存在）
		if userID, exists := c.Get("user_id"); exists {
			event.Interface("user_id", userID)
		}

		event.Msg("HTTP request")
	}
}

// GinRecovery Gin 恢复中间件
// 从 panic 中恢复并记录错误
func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get(RequestIDKey)

				globalLogger.Error().
					Str("request_id", requestID.(string)).
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Interface("panic", err).
					Stack().
					Msg("Panic recovered")

				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}

// ContextLogger 为 context 添加日志器
func ContextLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID, _ := c.Get(RequestIDKey)

		// 创建带 request_id 的日志器
		logger := globalLogger.With().
			Str("request_id", requestID.(string)).
			Logger()

		// 将日志器添加到 context
		ctx := logger.WithContext(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// LoggerFromContext 从 Gin Context 中获取日志器
func LoggerFromContext(c *gin.Context) *zerolog.Logger {
	logger := zerolog.Ctx(c.Request.Context())
	if logger != nil && logger.GetLevel() != zerolog.Disabled {
		return logger
	}
	return &globalLogger
}
