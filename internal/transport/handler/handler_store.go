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

type ServiceStoreInterface interface {
	CreateStore(ctx context.Context, actor policy.Actor, name string, groupID *int) (domain.StoreDetails, error)
	GetStore(ctx context.Context, actor policy.Actor, storeID int) (domain.StoreDetails, error)
	UpdateStore(ctx context.Context, actor policy.Actor, storeID int, updateStore domain.StoreUpdate) (domain.StoreDetails, error)
	DeleteStore(ctx context.Context, actor policy.Actor, storeID int) error
	ListStores(ctx context.Context, actor policy.Actor) ([]domain.StoreDetails, error)
	ListAllStores(ctx context.Context, actor policy.Actor) ([]domain.StoreDetails, error)
}

type StoreHandler struct {
	storeService ServiceStoreInterface
	validate     *validator.Validate
}

// CreateStoreHandler обрабатывает создание нового магазина.
//
// @Security BearerAuth
// @Summary Create store
// @Description Create a new store (user's group or specified group for admin)
// @Tags stores
// @Accept json
// @Produce json
// @Param request body dto.StoreRequest true "store payload"
// @Success 201 {object} dto.StoreResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/stores [post]
func (h StoreHandler) CreateStoreHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Декодирование тела запроса
	var req dto.StoreRequest
	if err := helpers.DecodeJSON(w, r, logger, &req); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Валидация входных данных
	if err := h.validate.Struct(req); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 4. Вызов сервиса для создания магазина
	store, err := h.storeService.CreateStore(ctx, actor, req.Name, req.GroupID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"name": req.Name, "groupID": req.GroupID})
		return
	}

	// 5. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusCreated, dto.ToStoreResponse(store))
}

// GetStoreHandler возвращает магазин по ID.
//
// @Security BearerAuth
// @Summary Get store
// @Description Get store by ID
// @Tags stores
// @Produce json
// @Param id path int true "store ID"
// @Success 200 {object} dto.StoreResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/stores/{id} [get]
func (h StoreHandler) GetStoreHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Извлечение и парсинг ID из пути
	storeID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Вызов сервиса для получения магазина
	store, err := h.storeService.GetStore(ctx, actor, storeID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"storeId": storeID})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToStoreResponse(store))
}

// UpdateStoreHandler обновляет название магазина по ID.
//
// @Security BearerAuth
// @Summary Update store
// @Description Update store name
// @Tags stores
// @Produce json
// @Param id path int true "store ID"
// @Param request body dto.StoreUpdateRequest true "store payload"
// @Success 200 {object} dto.StoreResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/stores/{id} [patch]
func (h StoreHandler) UpdateStoreHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Парсинг ID
	storeID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Декодирование тела запроса
	var req dto.StoreUpdateRequest
	if err := helpers.DecodeJSON(w, r, logger, &req); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 4. Валидация нового имени (если передано)
	if err := h.validate.Struct(req); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 5. Вызов сервиса для обновления
	store, err := h.storeService.UpdateStore(ctx, actor, storeID, dto.ToStoreUpdateRequest(req))
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"storeId": storeID})
		return
	}

	// 6. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToStoreResponse(store))
}

// DeleteStoreHandler удаляет магазин по ID.
//
// @Security BearerAuth
// @Summary Delete store
// @Description Delete store by ID
// @Tags stores
// @Produce json
// @Param id path int true "store ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/stores/{id} [delete]
func (h StoreHandler) DeleteStoreHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Парсинг ID
	storeID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Вызов сервиса для удаления
	if err := h.storeService.DeleteStore(ctx, actor, storeID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"storeId": storeID})
		return
	}

	// 4. Успешное удаление без контента
	w.WriteHeader(http.StatusNoContent)
}

// ListStoresHandler возвращает список магазинов, доступных пользователю (из его группы).
//
// @Security BearerAuth
// @Summary List user's stores
// @Description Get list of stores belonging to user's group
// @Tags stores
// @Produce json
// @Success 200 {array} dto.StoreResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/stores [get]
func (h StoreHandler) ListStoresHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Вызов сервиса для получения списка
	list, err := h.storeService.ListStores(ctx, actor)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 3. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToStoreListResponse(list))
}

// ListAllStoresHandler возвращает список всех магазинов (только для администраторов).
//
// @Security BearerAuth
// @Summary List all stores (admin only)
// @Description Get list of all stores (requires admin role)
// @Tags stores
// @Produce json
// @Success 200 {array} dto.StoreResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/stores/all [get]
func (h StoreHandler) ListAllStoresHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Вызов сервиса для получения списка всех магазинов
	list, err := h.storeService.ListAllStores(ctx, actor)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 3. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToStoreListResponse(list))
}
