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

func (h UserHandler) GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)

	actor, ok := authctx.ActorFromContext(ctx)
	if !ok {
		return
	}

	userIDParam, err := helpers.ParsePositiveIntParam(r, "userId")
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

	helpers.WriteJSON(w, logger, http.StatusCreated, resp)
}

func (h UserHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)

	actor, ok := authctx.ActorFromContext(ctx)
	if !ok {
		return
	}

	userIDParam, err := helpers.ParsePositiveIntParam(r, "userId")
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

func (h UserHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)

	actor, ok := authctx.ActorFromContext(ctx)
	if !ok {
		return
	}

	userIDParam, err := helpers.ParsePositiveIntParam(r, "userId")
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

	helpers.WriteJSON(w, logger, http.StatusCreated, resp)
}

func (h UserHandler) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)

	actor, ok := authctx.ActorFromContext(ctx)
	if !ok {
		return
	}

	user, err := h.userService.ListUsers(ctx, actor)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}
	resp := dto.ToUsersResponse(user)

	helpers.WriteJSON(w, logger, http.StatusCreated, resp)
}
