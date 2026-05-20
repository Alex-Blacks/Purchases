package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Alex-Blacks/Purchases/internal/logging"
)

type wrapped struct {
	w       http.ResponseWriter
	status  int
	written int
}

func (wr *wrapped) Header() http.Header {
	return wr.w.Header()
}

func (wr *wrapped) WriteHeader(code int) {
	wr.status = code
	wr.w.WriteHeader(code)
}
func (wr *wrapped) Write(b []byte) (int, error) {
	if wr.status == 0 {
		wr.status = http.StatusOK
	}
	n, err := wr.w.Write(b)
	wr.written += n
	return n, err
}

func LoggingMiddleware(baseLoger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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

			wrappedWriter := &wrapped{w: w}
			start := time.Now()
			next.ServeHTTP(wrappedWriter, r)
			duration := time.Since(start)
			log.Info("request finished",
				"status", wrappedWriter.status,
				"bytes", wrappedWriter.written,
				"duration", duration,
			)
		})
	}

}
