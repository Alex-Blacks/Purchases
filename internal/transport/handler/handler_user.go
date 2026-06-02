package handler

import (
	"context"
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/authctx"
	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
)

type ServiceUserInterface interface {
	CreateUser(ctx context.Context, name string, password string, email string, role string, status string) (domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	GetUserByID(ctx context.Context, actor policy.Actor, userID int) (domain.User, error)
	DeleteUser(ctx context.Context, actor policy.Actor, userID int) error
	UpdateUser(ctx context.Context, actor policy.Actor, userID int, updateUser domain.UpdateUser) (domain.User, error)
	ListUsers(ctx context.Context, actor policy.Actor) ([]domain.User, error)

	CheckPassword(user domain.User, password string) error
	GeneratePassword(password string) (string, error)

	GetAccessibleUser(ctx context.Context, actor policy.Actor, userID int) (domain.User, error)
}

type UserHandler struct {
	userService ServiceUserInterface
}

// CreateUserHandler godoc
//
// @Summary Create user
// @Description Create user
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.UserRequest true "user payload"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/users [post]
func (h UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)

	var req dto.UserRequest

	if err := helpers.DecodeJSON(w, r, logger, &req); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}
	if err := helpers.ValidateCreateUser(req); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	const (
		role   string = "user"
		status string = "active"
	)

	user, err := h.userService.CreateUser(ctx, req.Name, req.Password, req.Email, role, status)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{
			"name":  req.Name,
			"email": req.Email,
		})
		return
	}

	resp := dto.ToUserResponse(user)

	helpers.WriteJSON(w, logger, http.StatusCreated, resp)
}

// GetUserByIDHandler godoc
//
// @Security BearerAuth
// @Summary Get user by ID
// @Description Get user by ID
// @Tags users
// @Produce json
// @Param id path int true "user ID"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/users/{id} [get]
func (h UserHandler) GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)

	actor, ok := authctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	userIDParam, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.userService.GetUserByID(ctx, actor, userIDParam)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{
			"userIdParam": userIDParam,
		})
		return
	}
	resp := dto.ToUserResponse(user)

	helpers.WriteJSON(w, logger, http.StatusOK, resp)
}

// DeleteUserHandler godoc
//
// @Security BearerAuth
// @Summary delete user by ID
// @Description delete user by ID
// @Tags users
// @Produce json
// @Param id path int true "user ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/users/{id} [delete]
func (h UserHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)

	actor, ok := authctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	userIDParam, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.userService.DeleteUser(ctx, actor, userIDParam); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{
			"userIdParam": userIDParam,
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateUserHandler godoc
//
// @Security BearerAuth
// @Summary Update user
// @Description Update user
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "user ID"
// @Param request body dto.UserUpdateRequest true "user payload"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/users/{id} [patch]
func (h UserHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)

	actor, ok := authctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	userIDParam, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	var req dto.UserUpdateRequest

	if err := helpers.DecodeJSON(w, r, logger, &req); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.userService.UpdateUser(ctx, actor, userIDParam, dto.ToUserUpdateRequest(req))
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{
			"userIdParam": userIDParam,
			"name":        req.Name,
			"email":       req.Email,
			"roleRequest": req.Role,
			"status":      req.Status,
		})
		return
	}
	resp := dto.ToUserResponse(user)

	helpers.WriteJSON(w, logger, http.StatusOK, resp)
}

// ListUsersHandler godoc
//
// @Security BearerAuth
// @Summary list users
// @Description list users
// @Tags users
// @Produce json
// @Success 200 {array} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/users [get]
func (h UserHandler) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)

	actor, ok := authctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.userService.ListUsers(ctx, actor)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}
	resp := dto.ToUsersResponse(user)

	helpers.WriteJSON(w, logger, http.StatusOK, resp)
}
