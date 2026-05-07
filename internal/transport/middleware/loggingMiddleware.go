package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Alex-Blacks/Purchases/pkg"
)

func LoggingMiddleware(next http.Handler, baseLoger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestID := RequestIDFromContext(r.Context())
		if requestID == "" {
			requestID = "unknown"
		}

		logger := baseLoger.With(
			"request-id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
		)

		logger.Info("request started")

		ctx := pkg.WithContext(r.Context(), logger)
		r = r.WithContext(ctx)

		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		logger.Info("request succeeded",
			"duration", duration,
		)
	})
}
