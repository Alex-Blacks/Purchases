package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/Alex-Blacks/Purchases/internal/actorctx"
	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"

	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logging.LoggerFromContext(r.Context())

			authHeader := strings.Split(r.Header.Get("Authorization"), " ")
			if len(authHeader) != 2 || strings.ToLower(authHeader[0]) != "bearer" {
				logger.Warn("missing or malformed Authorization header")
				helpers.WriteError(w, logger, http.StatusUnauthorized, "требуется авторизация")
				return
			}

			tokenString := authHeader[1]

			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
				return []byte(secret), nil
			})
			if err != nil {
				logger.Warn("failed to parse JWT", "error", err)
				helpers.WriteError(w, logger, http.StatusUnauthorized, "неверный токен")
				return
			}

			if !token.Valid {
				logger.Warn("invalid JWT token")
				helpers.WriteError(w, logger, http.StatusUnauthorized, "неверный токен")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				logger.Warn("invalid JWT claims type")
				helpers.WriteError(w, logger, http.StatusUnauthorized, "неверный токен")
				return
			}

			expFloat, ok := claims["exp"].(float64)
			if !ok {
				logger.Warn("missing exp claim")
				helpers.WriteError(w, logger, http.StatusUnauthorized, "неверный токен")
				return
			}
			if int64(expFloat) <= time.Now().Unix() {
				logger.Warn("JWT expired")
				helpers.WriteError(w, logger, http.StatusUnauthorized, "токен истек")
				return
			}

			userIDFloat, ok1 := claims["sub"].(float64)
			groupIDFloat, ok2 := claims["group"].(float64)
			role, ok3 := claims["role"].(string)
			if !ok1 || !ok2 || !ok3 {
				logger.Warn("missing userID, groupID or role in JWT claims")
				helpers.WriteError(w, logger, http.StatusUnauthorized, "неверный токен")
				return
			}

			actor := policy.ToActor(int(userIDFloat), int(groupIDFloat), domain.UserRole(role))

			logger = logger.With("actor", actor)

			ctx := logging.WithContext(r.Context(), logger)
			ctx = actorctx.WithActor(ctx, actor)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

}
