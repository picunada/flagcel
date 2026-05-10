package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

func RequestID(base *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-ID")
			if id == "" {
				genuuid, _ := uuid.NewV7()
				id = genuuid.String()
			}

			logger := base.With(slog.String("request_id", id))
			ctx := context.WithValue(r.Context(), "logger", logger)

			w.Header().Set("X-Request-ID", id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
