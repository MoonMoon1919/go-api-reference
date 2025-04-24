package middleware

import (
	"log/slog"
	"net/http"

	"github.com/moonmoon1919/go-api-reference/internal/responses"
)

const (
	msgInternalServerError = "INTERNAL_SERVER_ERROR"
	keyError               = "ERROR"
)

type errorMiddleware struct {
	next http.Handler
}

func (m *errorMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handlePanics(w, r, m.next)
}

func ErrorHandlingMiddleware(next http.Handler) http.Handler {
	return &errorMiddleware{next: next}
}

/*
Handles panics and logs them as internal server errors
*/
func handlePanics(w http.ResponseWriter, r *http.Request, next http.Handler) {
	defer func() {
		if err := recover(); err != nil {
			slog.LogAttrs(r.Context(), slog.LevelError, msgInternalServerError, slog.Any(keyError, err))
			responses.WriteInternalServerErrorResponse(w)
			return
		}
	}()

	next.ServeHTTP(w, r)
}
