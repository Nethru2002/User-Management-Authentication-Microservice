package middleware

import (
	"net/http"
	"time"

	"auth-service/internal/infrastructure/metrics"
)

func MetricsMiddleware(m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m.ActiveRequests.Inc()
			defer m.ActiveRequests.Dec()

			start := time.Now()
			wrapped := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			m.TrackRequest(r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
		})
	}
}