package helpers

import (
	"log/slog"
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/authctx"
)

// CheckAdminPermission проверяет, что роль пользователя admin
func CheckAdminPermission(w http.ResponseWriter, r *http.Request, logger *slog.Logger) bool {
	ctx := r.Context()
	role, ok := authctx.RoleFromContext(ctx)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return false
	}
	if role != "admin" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return false
	}
	return true
}

// CheckUserOrAdminPermission проверяет, что либо сам пользователь, либо admin
func CheckUserOrAdminPermission(w http.ResponseWriter, r *http.Request, logger *slog.Logger, userIDParam int) bool {
	ctx := r.Context()
	userID, ok1 := authctx.UserIDFromContext(ctx)
	role, ok2 := authctx.RoleFromContext(ctx)
	if !ok1 || !ok2 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return false
	}
	if userID != userIDParam && role != "admin" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return false
	}
	return true
}
