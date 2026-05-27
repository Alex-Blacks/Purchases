package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/Alex-Blacks/Purchases/internal/authctx"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"

	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logging.LoggerFromContext(r.Context())

			authHeader := strings.Split(r.Header.Get("Authorization"), " ")
			if len(authHeader) != 2 || strings.ToLower(authHeader[0]) != "bearer" {
				logger.Warn("missing or malformed Authorization header")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			tokenString := authHeader[1]

			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
				return []byte(secret), nil
			})
			if err != nil {
				logger.Warn("failed to parse JWT", "error", err)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				logger.Warn("invalid JWT token")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				logger.Warn("invalid JWT claims type")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			expFloat, ok := claims["exp"].(float64)
			if !ok {
				logger.Warn("missing exp claim")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if int64(expFloat) <= time.Now().Unix() {
				logger.Warn("JWT expired")
				http.Error(w, "token expired", http.StatusUnauthorized)
				return
			}

			userIDFloat, ok1 := claims["sub"].(float64)
			role, ok2 := claims["role"].(string)
			if !ok1 || !ok2 {
				logger.Warn("missing userID or role in JWT claims")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			userID := int(userIDFloat)
			actor := policy.ToActor(userID, policy.Role(role))

			logger = logger.With("actor", actor)

			ctx := logging.WithContext(r.Context(), logger)
			ctx = authctx.WithActor(ctx, actor)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

}
