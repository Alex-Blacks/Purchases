package handler

import (
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/authctx"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
)

func CreateUserHandler(svc *service.ServiceUser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		user, err := svc.CreateUser(ctx, req.Name, req.Password, req.Email, role, status)
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
}

func GetUserByIDHandler(svc *service.ServiceUser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		user, err := svc.GetUserByID(ctx, actor, userIDParam)
		if err != nil {
			helpers.WriteDomainError(w, logger, err, map[string]any{
				"userIdParam": userIDParam,
			})
			return
		}
		resp := dto.ToUserResponse(user)

		helpers.WriteJSON(w, logger, http.StatusCreated, resp)
	}
}

func DeleteUserHandler(svc *service.ServiceUser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		if err := svc.DeleteUser(ctx, actor, userIDParam); err != nil {
			helpers.WriteDomainError(w, logger, err, map[string]any{
				"userIdParam": userIDParam,
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func UpdateUserHandler(svc *service.ServiceUser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		user, err := svc.UpdateUser(ctx, actor, userIDParam, dto.ToUserUpdateRequest(req))
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
}

func ListUsersHandler(svc *service.ServiceUser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.LoggerFromContext(ctx)

		actor, ok := authctx.ActorFromContext(ctx)
		if !ok {
			return
		}

		user, err := svc.ListUsers(ctx, actor)
		if err != nil {
			helpers.WriteDomainError(w, logger, err, nil)
			return
		}
		resp := dto.ToUsersResponse(user)

		helpers.WriteJSON(w, logger, http.StatusCreated, resp)
	}
}
