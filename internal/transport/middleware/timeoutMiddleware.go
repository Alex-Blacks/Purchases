package middleware

import (
	"context"
	"net/http"
	"time"
)

func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if deadline, ok := r.Context().Deadline(); ok {
				remaining := time.Until(deadline)
				if remaining <= 0 {
					http.Error(w, "request timeout", http.StatusRequestTimeout)
					return
				}
				if remaining <= timeout {
					next.ServeHTTP(w, r)
					return
				}
			}
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}
