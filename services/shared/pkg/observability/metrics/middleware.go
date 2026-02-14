package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// HTTPMiddleware возвращает middleware для сбора Prometheus метрик
func HTTPMiddleware(serviceName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Увеличиваем счетчик in-flight запросов
			HTTPRequestsInFlight.WithLabelValues(serviceName).Inc()
			defer HTTPRequestsInFlight.WithLabelValues(serviceName).Dec()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			duration := time.Since(start)
			status := strconv.Itoa(ww.Status())

			RecordHTTPMetrics(serviceName, r.Method, r.URL.Path, status, duration)
		})
	}
}
