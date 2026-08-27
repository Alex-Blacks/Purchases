package handler

import (
	"context"
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/actorctx"
	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
	"github.com/go-playground/validator/v10"
)

type ServiceGroupInterface interface {
	CreateGroup(ctx context.Context, actor policy.Actor, name string, adminUserID int) (domain.GroupDetails, error)
	GetByID(ctx context.Context, actor policy.Actor, groupID int) (domain.GroupDetails, error)
	UpdateByID(ctx context.Context, actor policy.Actor, groupID int, updateGroup domain.GroupUpdate) (domain.GroupDetails, error)
	DeleteByID(ctx context.Context, actor policy.Actor, groupID int) error
	ListAll(ctx context.Context, actor policy.Actor) ([]domain.GroupDetails, error)
}

type GroupHandler struct {
	groupService ServiceGroupInterface
	validate     *validator.Validate
}

// CreateGroupHandler обрабатывает создание новой группы.
//
// @Security BearerAuth
// @Summary Create group
// @Description Create a new group
// @Tags groups
// @Accept json
// @Produce json
// @Param request body dto.GroupRequest true "group payload"
// @Success 201 {object} dto.GroupResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/groups [post]
func (h *GroupHandler) CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Декодирование и валидация тела запроса
	var req dto.GroupRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, &req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	// 3. Вызов сервиса
	group, err := h.groupService.CreateGroup(ctx, actor, req.Name, req.AdminUserID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"name": req.Name, "adminUserId": req.AdminUserID})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusCreated, dto.ToGroupResponse(group))
}

// GetGroupHandler возвращает группу по ID.
//
// @Security BearerAuth
// @Summary Get group
// @Description Get group by ID
// @Tags groups
// @Produce json
// @Param id path int true "group ID"
// @Success 200 {object} dto.GroupResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/groups/{id} [get]
func (h GroupHandler) GetGroupHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Извлечение и парсинг ID из пути
	groupID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Вызов сервиса для получения группы
	group, err := h.groupService.GetByID(ctx, actor, groupID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"groupId": groupID})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToGroupResponse(group))
}

// UpdateGroupHandler обновляет название группы по ID.
//
// @Security BearerAuth
// @Summary Update group
// @Description Update group name
// @Tags groups
// @Produce json
// @Param id path int true "group ID"
// @Param request body dto.GroupUpdateRequest true "group payload"
// @Success 200 {object} dto.GroupResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/groups/{id} [patch]
func (h GroupHandler) UpdateGroupHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Парсинг ID
	groupID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Декодирование и валидация тела запроса
	var req dto.GroupUpdateRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, &req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	// 4. Вызов сервиса для обновления
	group, err := h.groupService.UpdateByID(ctx, actor, groupID, dto.ToGroupUpdateRequest(req))
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"groupId": groupID})
		return
	}

	// 5. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToGroupResponse(group))
}

// DeleteGroupHandler удаляет группу по ID.
//
// @Security BearerAuth
// @Summary Delete group
// @Description Delete group by ID
// @Tags groups
// @Produce json
// @Param id path int true "group ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/groups/{id} [delete]
func (h GroupHandler) DeleteGroupHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Парсинг ID
	groupID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Вызов сервиса для удаления
	if err := h.groupService.DeleteByID(ctx, actor, groupID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"groupId": groupID})
		return
	}

	// 4. Успешное удаление без контента
	w.WriteHeader(http.StatusNoContent)
}

// ListAllGroupsHandler возвращает список всех групп (только для администраторов).
//
// @Security BearerAuth
// @Summary List all groups
// @Description Get list of groups
// @Tags groups
// @Produce json
// @Success 200 {array} dto.GroupResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/groups [get]
func (h GroupHandler) ListAllGroupsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Вызов сервиса для получения списка
	list, err := h.groupService.ListAll(ctx, actor)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 3. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToGroupListResponse(list))
}
