package handler

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/actorctx"
	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
	"github.com/go-playground/validator/v10"
)

type ServiceUserInterface interface {
	Create(ctx context.Context, name string, password string, email string, role domain.UserRole, status domain.UserStatus) (domain.UserDetails, error)
	GetByID(ctx context.Context, actor policy.Actor, userID int) (domain.UserDetails, error)
	GetByEmail(ctx context.Context, email string) (domain.UserDetails, error)
	UpdateByID(ctx context.Context, actor policy.Actor, userID int, updateUser domain.UserUpdate) (domain.UserDetails, error)
	DeleteByID(ctx context.Context, actor policy.Actor, userID int) error
	List(ctx context.Context, actor policy.Actor, filter domain.UserListFilter) ([]domain.UserDetails, error)
	Count(ctx context.Context, actor policy.Actor, filter domain.UserListFilter) (int, error)
}

type UserHandler struct {
	userService ServiceUserInterface
	validate    *validator.Validate
}

// CreateUserHandler обрабатывает создание нового пользователя.
//
// @Summary Create user
// @Description Create a new user (user's group or specified group for admin)
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.UserRequest true "user payload"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /users [post]
func (h UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	if !actor.HasRole(domain.RoleAdmin) {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Декодирование и валидация тела запроса
	var req dto.UserRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, &req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	// 3. Вызов сервиса для создания пользователя
	user, err := h.userService.Create(ctx, req.Name, req.Password, req.Email, *req.Role, domain.UserStatusActive)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{
			"name":       req.Name,
			"email_hash": fmt.Sprintf("%x", sha256.Sum256([]byte(req.Email))),
		})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusCreated, dto.ToUserResponse(user))
}

// GetUserByIDHandler возвращает пользователя по ID.
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
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/users/{id} [get]
func (h UserHandler) GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Извлечение и парсинг ID из пути
	userID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Вызов сервиса для получения пользователя
	user, err := h.userService.GetByID(ctx, actor, userID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{
			"userId": userID,
		})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToUserResponse(user))
}

// UpdateUserHandler обновляет пользователя по ID.
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
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/users/{id} [patch]
func (h UserHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Извлечение и парсинг ID из пути
	userID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Декодирование и валидация тела запроса
	var req dto.UserUpdateRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, &req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	// 4. Вызов сервиса для получения пользователя
	user, err := h.userService.UpdateByID(ctx, actor, userID, dto.ToUserUpdateRequest(req))
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{
			"userIdParam": userID,
			"name":        req.Name,
			"email":       req.Email,
			"roleRequest": req.Role,
			"status":      req.Status,
		})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToUserResponse(user))
}

// DeleteUserHandler удаляет пользователя по ID.
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
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/users/{id} [delete]
func (h UserHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Извлечение и парсинг ID из пути
	userID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Вызов сервиса для удаления пользователя
	if err := h.userService.DeleteByID(ctx, actor, userID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{
			"userId": userID,
		})
		return
	}

	// 4. Успешное удаление без контента
	w.WriteHeader(http.StatusNoContent)
}

// ListUsersHandler возвращает список пользователей с фильтрацией через query-параметры.
//
// @Security BearerAuth
// @Summary list users
// @Description list users with pagination and filters
// @Tags users
// @Produce json
// @Param group_ids[] query []int false "Group IDs"
// @Param role query string false "Role (admin/user)" Enums(admin, user)
// @Param status query string false "Status (active/blocked)" Enums(active, blocked)
// @Param created_from query string false "Created from (RFC3339)"
// @Param created_to query string false "Created to (RFC3339)"
// @Param updated_from query string false "Updated from (RFC3339)"
// @Param updated_to query string false "Updated to (RFC3339)"
// @Param limit query int false "Limit" default(10) minimum(1) maximum(100)
// @Param offset query int false "Offset" default(0) minimum(0)
// @Success 200 {array} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/users [get]
func (h UserHandler) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Биндим query-параметры в структуру
	var queryFilter dto.UserFilterQuery
	if err := helpers.FormDecoder.Decode(&queryFilter, r.URL.Query()); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, "invalid query parameters: "+err.Error())
		return
	}

	if err := h.validate.Struct(queryFilter); err != nil {
		helpers.WriteDomainError(w, logger, err, queryFilter)
		return
	}

	// 3. Вызываем сервис
	users, err := h.userService.List(ctx, actor, queryFilter.ToUserFilterRequest())
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 4. Отправляем ответ
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToUserListResponse(users))
}

// CountUsersHandler возвращает количество всех пользователей.
//
// @Security BearerAuth
// @Summary list users
// @Description list users
// @Tags users
// @Produce json
// @Param group_ids[] query []int false "Group IDs"
// @Param role query string false "Role (admin/user)" Enums(admin, user)
// @Param status query string false "Status (active/blocked)" Enums(active, blocked)
// @Param created_from query string false "Created from (RFC3339)"
// @Param created_to query string false "Created to (RFC3339)"
// @Param updated_from query string false "Updated from (RFC3339)"
// @Param updated_to query string false "Updated to (RFC3339)"
// @Param limit query int false "Limit" default(10) minimum(1) maximum(100)
// @Param offset query int false "Offset" default(0) minimum(0)
// @Success 200 {object} dto.CountResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/users/all [get]
func (h UserHandler) CountUsersHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Биндим query-параметры в структуру
	var queryFilter dto.UserFilterQuery
	if err := helpers.FormDecoder.Decode(&queryFilter, r.URL.Query()); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, "invalid query parameters: "+err.Error())
		return
	}

	if err := h.validate.Struct(queryFilter); err != nil {
		helpers.WriteDomainError(w, logger, err, queryFilter)
		return
	}

	// 3. Вызов сервиса для получения количества всех пользователей
	count, err := h.userService.Count(ctx, actor, queryFilter.ToUserFilterRequest())
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 3. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.CountResponse{Count: count})
}
