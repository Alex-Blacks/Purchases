package handler

import (
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
)

func CreateUserHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.LoggerFromContext(ctx)

		var req dto.UserRequest

		if !helpers.DecodeJSONHelper(w, r, logger, &req) {
			return
		}
		if err := helpers.ValidateStruct(req); err != nil {
			logger.Warn("invalid parameter", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		role := "user"
		status := "active"

		user, err := svc.CreateUser(ctx, req.Name, req.Password, req.Email, role, status)
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"name":  req.Name,
				"email": req.Email,
			})
			return
		}

		resp := dto.ToUserResponse(user)

		helpers.RespondJSON(w, http.StatusCreated, resp, logger)
	}
}

func GetUserByIDHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.LoggerFromContext(ctx)
		userIDParam, ok := helpers.ParsePositiveIntParam(w, r, "userId", logger)
		if !ok {
			return
		}

		if !helpers.CheckUserOrAdminPermission(w, r, logger, userIDParam) {
			return
		}

		user, err := svc.GetUserByID(ctx, userIDParam)
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"userIdParam": userIDParam,
			})
			return
		}
		resp := dto.ToUserResponse(user)

		helpers.RespondJSON(w, http.StatusOK, resp, logger)
	}
}

func DeleteUserHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.LoggerFromContext(ctx)
		userIDParam, ok := helpers.ParsePositiveIntParam(w, r, "userId", logger)
		if !ok {
			return
		}

		if !helpers.CheckUserOrAdminPermission(w, r, logger, userIDParam) {
			return
		}

		if err := svc.DeleteUser(ctx, userIDParam); err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"userIdParam": userIDParam,
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func UpdateUserHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.LoggerFromContext(ctx)
		userIDParam, ok := helpers.ParsePositiveIntParam(w, r, "userId", logger)
		if !ok {
			return
		}

		var req dto.UserUpdateRequest

		if !helpers.DecodeJSONHelper(w, r, logger, &req) {
			return
		}

		if !helpers.CheckUserOrAdminPermission(w, r, logger, userIDParam) {
			return
		}
		if req.Role == "admin" {
			if !helpers.CheckAdminPermission(w, r, logger) {
				return
			}
		}

		user, err := svc.UpdateUser(ctx, userIDParam, dto.ToUserUpdateRequest(req))
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"userIdParam": userIDParam,
				"name":        req.Name,
				"email":       req.Email,
				"roleRequest": req.Role,
			})
			return
		}
		resp := dto.ToUserResponse(user)

		helpers.RespondJSON(w, http.StatusOK, resp, logger)
	}
}

func ListUsersHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.LoggerFromContext(ctx)

		if !helpers.CheckAdminPermission(w, r, logger) {
			return
		}

		user, err := svc.ListUsers(ctx)
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{})
			return
		}
		resp := dto.ToUsersResponse(user)

		helpers.RespondJSON(w, http.StatusOK, resp, logger)
	}
}
