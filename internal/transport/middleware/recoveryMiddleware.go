package middleware

import (
	"encoding/json"
	"net/http"
	"runtime/debug"

	"github.com/Alex-Blacks/Purchases/internal/logging"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			logger := logging.LoggerFromContext(r.Context())

			if rec := recover(); rec != nil {
				logger.Error("panic",
					"error", rec,
					"stack", debug.Stack(),
				)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(map[string]string{"Error": "internal server error"}); err != nil {
					logger.Error("failed to encode recovery response", "error", err)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
