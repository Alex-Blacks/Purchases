package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
)

func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := logging.LoggerFromContext(ctx)
			if deadline, ok := ctx.Deadline(); ok {
				remaining := time.Until(deadline)
				if remaining <= 0 {
					helpers.WriteError(w, logger, http.StatusRequestTimeout, "время ожидания запроса истекло")
					return
				}
				if remaining <= timeout {
					next.ServeHTTP(w, r)
					return
				}
			}
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}
