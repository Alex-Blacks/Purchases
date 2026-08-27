package handler

import (
	"context"
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
	"github.com/go-playground/validator/v10"
)

type ServiceAuthInterface interface {
	Login(ctx context.Context, email, password string) (domain.Login, error)
	Register(ctx context.Context, name, email, password string) (domain.Login, error)
}
type AuthHandler struct {
	authService ServiceAuthInterface
	validate    *validator.Validate
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
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /login [post]
func (h AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)

	// 2. Декодирование и валидация тела запроса
	var req dto.LoginRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, &req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	// 3. Вызов сервиса авторизации
	result, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"email": req.Email})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToLoginResponse(result))
}

// RegisterHandler godoc
//
// @Summary Register
// @Description Register
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "register payload"
// @Success 200 {object} dto.RegisterResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /register [post]
func (h AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)

	// 2. Декодирование и валидация тела запроса
	var req dto.RegisterRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, &req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	// 3. Вызов сервиса регистрации
	result, err := h.authService.Register(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"name": req.Name, "email": req.Email})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToRegisterResponse(result))
}
