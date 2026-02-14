package logging

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// HTTPMiddleware - Chi middleware для логирования HTTP запросов
func HTTPMiddleware(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Получаем request_id из chi middleware (если есть)
			requestID := middleware.GetReqID(r.Context())
			if requestID != "" {
				ctx := ContextWithRequestID(r.Context(), requestID)
				r = r.WithContext(ctx)
			}

			// Wrap response writer для захвата статуса
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Обработка запроса
			next.ServeHTTP(ww, r)

			// Логирование после обработки
			duration := time.Since(start)

			// Создаем logger с trace context
			loggerWithContext := WithTraceContext(r.Context(), logger)

			loggerWithContext.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("query", r.URL.RawQuery),
				zap.Int("status", ww.Status()),
				zap.Int("bytes", ww.BytesWritten()),
				zap.Duration("latency", duration),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			)
		})
	}
}
