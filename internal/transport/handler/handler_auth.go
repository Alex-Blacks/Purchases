package handler

import (
	"context"
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
)

type ServiceAuthInterface interface {
	Login(ctx context.Context, email, password string) (string, error)
}
type AuthHandler struct {
	authService ServiceAuthInterface
}

// LoginHandler godoc
//
// @Summary Login
// @Description Login
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "login payload"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /login [post]
func (h AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	var req dto.LoginRequest

	if err := helpers.DecodeJSON(w, r, logger, &req); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"email": req.Email})
		return
	}

	resp := dto.LoginResponse{Token: token}
	helpers.WriteJSON(w, logger, http.StatusOK, resp)
}
