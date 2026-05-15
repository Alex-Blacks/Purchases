package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Alex-Blacks/Purchases/internal/logging"
)

func LoggingMiddleware(next http.Handler, baseLoger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestID, ok := RequestIDFromContext(r.Context())
		if requestID == "" || !ok {
			requestID = "unknown"
		}

		log := baseLoger.With(
			"request-id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
		)

		log.Info("request started")
		ctx := logging.WithContext(r.Context(), log)
		r = r.WithContext(ctx)

		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		log.Info("request succeeded",
			"duration", duration,
		)
	})
}
