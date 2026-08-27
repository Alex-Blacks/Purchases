package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
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

				helpers.WriteInternalError(w, logger, fmt.Errorf("panic: %v", rec), nil)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
