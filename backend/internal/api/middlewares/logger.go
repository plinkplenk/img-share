package middlewares

import (
	"log/slog"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	code int
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loggerWriter := &loggingResponseWriter{ResponseWriter: w}
			timeBefore := time.Now()
			next.ServeHTTP(loggerWriter, r)
			logger.Info(
				"incoming request",
				"method", r.Method,
				"uri", r.RequestURI,
				"code", loggerWriter.code,
				"duration", time.Since(timeBefore).Milliseconds(),
			)
		})
	}
}
