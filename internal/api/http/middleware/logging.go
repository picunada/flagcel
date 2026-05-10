package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"
)

func Logging(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := httptest.NewRecorder()

			next.ServeHTTP(rec, r)

			// Copy buffered response back to the real writer.
			for k, v := range rec.Header() {
				w.Header()[k] = v
			}
			w.WriteHeader(rec.Code)
			rec.Body.WriteTo(w)

			logger.Info("http_request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.Code),
				slog.Int("bytes", rec.Body.Len()),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}
