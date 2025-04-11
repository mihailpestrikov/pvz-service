package middleware

import (
	"net/http"
	"strconv"
	"time"

	"avito-backend-trainee-assignment-spring-2025/pkg/metrics"
)

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		respWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(respWriter, r)

		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(respWriter.statusCode)

		metrics.RequestsTotal.WithLabelValues(r.Method, r.URL.Path, statusCode).Inc()
		metrics.RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
