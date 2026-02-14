package logging

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GinMiddleware - Gin middleware для логирования HTTP запросов
func GinMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Обработка запроса
		c.Next()

		// Логирование после обработки
		duration := time.Since(start)

		// Извлечение request_id если есть
		requestID, exists := c.Get("request_id")
		if !exists {
			requestID = ""
		}

		// Создаем контекст с request_id
		ctx := c.Request.Context()
		if reqID, ok := requestID.(string); ok && reqID != "" {
			ctx = ContextWithRequestID(ctx, reqID)
		}

		// Создаем logger с trace context
		loggerWithContext := WithTraceContext(ctx, logger)

		// Логирование
		loggerWithContext.Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.Int("status", c.Writer.Status()),
			zap.Int("bytes", c.Writer.Size()),
			zap.Duration("latency", duration),
			zap.String("remote_addr", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)
	}
}
