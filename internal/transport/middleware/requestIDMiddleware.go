package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type requestIDKeyType string

const requestIDKeyContext requestIDKeyType = "request-id"

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestID := r.Header.Get("X-Request-ID")
		_, err := uuid.Parse(requestID)
		if err != nil || requestID == "" {
			requestID = uuid.NewString()
		}

		ctx := context.WithValue(r.Context(), requestIDKeyContext, requestID)
		r = r.WithContext(ctx)

		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r)

	})
}

func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKeyContext).(string); ok {
		return id
	}
	return ""
}
