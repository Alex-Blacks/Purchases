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
	Create(ctx context.Context, actor policy.Actor, params domain.StoreCreate, groupID *int) (domain.StoreDetails, error)
	Get(ctx context.Context, actor policy.Actor, id int) (domain.StoreDetails, error)
	Update(ctx context.Context, actor policy.Actor, id int, updates domain.StoreUpdate) (domain.StoreDetails, error)
	Delete(ctx context.Context, actor policy.Actor, id int) error
	List(ctx context.Context, actor policy.Actor, filter domain.StoreListFilter) ([]domain.StoreDetails, error)
	Count(ctx context.Context, actor policy.Actor, filter domain.StoreListFilter) (int, error)
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

	// 2. Декодирование и валидация тела запроса
	var req dto.StoreRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, &req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	// 3. Вызов сервиса для создания магазина
	params := domain.StoreCreate{Name: req.Name}
	store, err := h.storeService.Create(ctx, actor, params, req.GroupID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"params": params, "groupID": req.GroupID})
		return
	}

	// 4. Формирование и отправка ответа
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
	store, err := h.storeService.Get(ctx, actor, storeID)
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

	// 3. Декодирование и валидация тела запроса
	var req dto.StoreUpdateRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, &req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	// 4. Вызов сервиса для обновления
	store, err := h.storeService.Update(ctx, actor, storeID, dto.ToStoreUpdateRequest(req))
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"storeId": storeID})
		return
	}

	// 5. Формирование и отправка ответа
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
	if err := h.storeService.Delete(ctx, actor, storeID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"storeId": storeID})
		return
	}

	// 4. Успешное удаление без контента
	w.WriteHeader(http.StatusNoContent)
}

// ListStoresHandler возвращает список магазинов.
//
// @Security BearerAuth
// @Summary List user's stores
// @Description Get list of stores belonging to user's group
// @Tags stores
// @Produce json
// @Param group_ids[] query []int false "Group IDs"
// @Param name query string false "Name" minimum(1) maximum(50)
// @Param limit query int false "Limit" default(10) minimum(1) maximum(100)
// @Param offset query int false "Offset" default(0) minimum(0)
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

	// 2. Биндим query-параметры в структуру
	var queryFilter dto.StoreFilterQuery
	if err := helpers.FormDecoder.Decode(&queryFilter, r.URL.Query()); err != nil {
		logger.WarnContext(ctx, "invalid query parameters", "error", err)
		helpers.WriteError(w, logger, http.StatusBadRequest, "недопустимые параметры запроса")
		return
	}

	if err := h.validate.Struct(queryFilter); err != nil {
		helpers.WriteDomainError(w, logger, err, queryFilter)
		return
	}

	// 3. Вызов сервиса для получения списка
	list, err := h.storeService.List(ctx, actor, queryFilter.ToDomainFilter())
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 4. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToStoreListResponse(list))
}

// CountStoresHandler возвращает количество всех магазинов.
//
// @Security BearerAuth
// @Summary Count all stores
// @Description Get count of all stores
// @Tags stores
// @Produce json
// @Param group_ids[] query []int false "Group IDs"
// @Param name query string false "Name" minimum(1) maximum(50)
// @Param limit query int false "Limit" default(10) minimum(1) maximum(100)
// @Param offset query int false "Offset" default(0) minimum(0)
// @Success 200 {object} dto.CountResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/stores/count [get]
func (h StoreHandler) CountStoresHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Биндим query-параметры в структуру
	var queryFilter dto.StoreFilterQuery
	if err := helpers.FormDecoder.Decode(&queryFilter, r.URL.Query()); err != nil {
		logger.WarnContext(ctx, "invalid query parameters", "error", err)
		helpers.WriteError(w, logger, http.StatusBadRequest, "недопустимые параметры запроса")
		return
	}

	if err := h.validate.Struct(queryFilter); err != nil {
		helpers.WriteDomainError(w, logger, err, queryFilter)
		return
	}

	// 3. Вызов сервиса для получения количества всех магазинов
	count, err := h.storeService.Count(ctx, actor, queryFilter.ToDomainFilter())
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 3. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.CountResponse{Count: count})
}
