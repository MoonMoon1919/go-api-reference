package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

const (
	msgRequestCompleted = "REQUEST_COMPLETED"
	keyMethod           = "method"
	keyStatus           = "status"
	keyRemoteAddr       = "remote_addr"
	keyUrl              = "url"
	keyDurationMs       = "duration_ms"
)

/*
A wrapper around the http.ResponseWriter that allows us to log the status code
*/
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(status int) {
	w.statusCode = status
	w.ResponseWriter.WriteHeader(status)
}

type loggingMiddleware struct {
	next http.Handler
}

func (m *loggingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	rw := &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	m.next.ServeHTTP(rw, r)

	slog.LogAttrs(r.Context(), slog.LevelInfo, msgRequestCompleted,
		slog.String(keyMethod, r.Method),
		slog.Int(keyStatus, rw.statusCode),
		slog.String(keyRemoteAddr, r.RemoteAddr),
		slog.String(keyUrl, r.URL.String()),
		slog.Int64(keyDurationMs, time.Since(start).Milliseconds()),
	)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return &loggingMiddleware{next: next}
}
