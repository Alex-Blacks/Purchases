package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Alex-Blacks/Purchases/internal/actorctx"
	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
	"github.com/go-playground/validator/v10"
)

type ServiceProductInterface interface {
	Create(ctx context.Context, actor policy.Actor, params domain.ProductCreate, groupID *int) (domain.ProductDetails, error)
	Get(ctx context.Context, actor policy.Actor, id int) (domain.ProductDetails, error)
	Update(ctx context.Context, actor policy.Actor, id int, updates domain.ProductUpdate) (domain.ProductDetails, error)
	Delete(ctx context.Context, actor policy.Actor, id int) error
	List(ctx context.Context, actor policy.Actor, filter domain.ProductListFilter) ([]domain.ProductDetails, error)
	Count(ctx context.Context, actor policy.Actor, filter domain.ProductListFilter) (int, error)
}

type ProductHandler struct {
	productService ServiceProductInterface
	validate       *validator.Validate
}

// CreateProductHandler обрабатывает создание нового продукта.
//
// @Security BearerAuth
// @Summary Create product
// @Description Create a new product (user's group or specified group for admin)
// @Tags products
// @Accept json
// @Produce json
// @Param request body dto.ProductRequest true "product payload"
// @Success 201 {object} dto.ProductResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/products [post]
func (h ProductHandler) CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Декодирование и валидация тела запроса
	var req dto.ProductRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, &req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	// 3. Вызов сервиса для создания продукта
	params := domain.ProductCreate{Title: req.Title}
	product, err := h.productService.Create(ctx, actor, params, req.GroupID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"params": params, "groupID": req.GroupID})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusCreated, dto.ToProductResponse(product))
}

// GetProductHandler возвращает продукт по ID.
//
// @Security BearerAuth
// @Summary Get product
// @Description Get product by ID
// @Tags products
// @Produce json
// @Param id path int true "product ID"
// @Success 200 {object} dto.ProductResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/products/{id} [get]
func (h ProductHandler) GetProductHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Извлечение и парсинг ID из пути
	productID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Вызов сервиса для получения продукта
	product, err := h.productService.Get(r.Context(), actor, productID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"productId": productID})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToProductResponse(product))
}

// UpdateProductHandler обновляет продукт по ID.
//
// @Security BearerAuth
// @Summary Update product
// @Description Update product name
// @Tags products
// @Produce json
// @Param id path int true "product ID"
// @Param request body dto.ProductUpdateRequest true "product payload"
// @Success 200 {object} dto.ProductResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/products/{id} [patch]
func (h ProductHandler) UpdateProductHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Парсинг ID
	productID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Декодирование и валидация тела запроса
	var req dto.ProductUpdateRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, &req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	// 4. Вызов сервиса для обновления
	product, err := h.productService.Update(ctx, actor, productID, dto.ToProductUpdateRequest(req))
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"productId": productID})
		return
	}

	// 5. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToProductResponse(product))
}

// DeleteProductHandler удаляет продукт по ID.
//
// @Security BearerAuth
// @Summary Delete product
// @Description Delete product by ID
// @Tags products
// @Produce json
// @Param id path int true "product ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/products/{id} [delete]
func (h ProductHandler) DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Парсинг ID
	productID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Вызов сервиса для удаления
	if err := h.productService.Delete(ctx, actor, productID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"productId": productID})
		return
	}

	// 4. Успешное удаление без контента
	w.WriteHeader(http.StatusNoContent)
}

// ListProductsHandler возвращает список продуктов.
//
// @Security BearerAuth
// @Summary List user's products
// @Description Get list of products belonging to user's group
// @Tags products
// @Produce json
// @Param group_ids[] query []int false "Group IDs"
// @Param title query string false "Title" minimum(1) maximum(50)
// @Param limit query int false "Limit" default(10) minimum(1) maximum(100)
// @Param offset query int false "Offset" default(0) minimum(0)
// @Success 200 {array} dto.ProductResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/products [get]
func (h ProductHandler) ListProductsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Биндим query-параметры в структуру
	var queryFilter dto.ProductFilterQuery
	if err := helpers.FormDecoder.Decode(&queryFilter, r.URL.Query()); err != nil {
		logger.WarnContext(ctx, "invalid query parameters: %w", err)
		helpers.WriteError(w, logger, http.StatusBadRequest, "недопустимые параметры запроса")
		return
	}

	if err := h.validate.Struct(queryFilter); err != nil {
		helpers.WriteDomainError(w, logger, err, queryFilter)
		return
	}

	// 3. Вызов сервиса для получения списка
	list, err := h.productService.List(ctx, actor, queryFilter.ToDomainFilter())
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 4. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToProductListResponse(list))
}

// CountProductsHandler возвращает список всех продуктов (только для администраторов).
//
// @Security BearerAuth
// @Summary Count all products
// @Description Get count of all products
// @Tags products
// @Produce json
// @Param group_ids[] query []int false "Group IDs"
// @Param title query string false "Title" minimum(1) maximum(50)
// @Param limit query int false "Limit" default(10) minimum(1) maximum(100)
// @Param offset query int false "Offset" default(0) minimum(0)
// @Success 200 {object} dto.CountResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/products/count [get]
func (h ProductHandler) CountProductsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Биндим query-параметры в структуру
	var queryFilter dto.ProductFilterQuery
	if err := helpers.FormDecoder.Decode(&queryFilter, r.URL.Query()); err != nil {
		logger.WarnContext(ctx, "invalid query parameters: %w", err)
		helpers.WriteError(w, logger, http.StatusBadRequest, "недопустимые параметры запроса")
		return
	}

	if err := h.validate.Struct(queryFilter); err != nil {
		helpers.WriteDomainError(w, logger, err, queryFilter)
		return
	}

	// 3. Вызов сервиса для получения списка всех продуктов
	count, err := h.productService.Count(ctx, actor, queryFilter.ToDomainFilter())
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 4. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.CountResponse{Count: count})
}

// ---------------------------------------------------------------------------------
// --------------------------------------ALIAS--------------------------------------
// ---------------------------------------------------------------------------------

type ServiceProductAliasInterface interface {
	Create(ctx context.Context, actor policy.Actor, productID int, alias string, groupID *int) (domain.ProductAliasDetails, error)
	GetByID(ctx context.Context, actor policy.Actor, aliasID int) (domain.ProductAliasDetails, error)
	UpdateByID(ctx context.Context, actor policy.Actor, aliasID int, updates domain.ProductAliasUpdate) (domain.ProductAliasDetails, error)
	DeleteByID(ctx context.Context, actor policy.Actor, aliasID int) error
	DeleteAllProductAliases(ctx context.Context, actor policy.Actor, productID int) error
	List(ctx context.Context, actor policy.Actor, filter domain.ProductAliasListFilter) ([]domain.ProductAliasDetails, error)
	Count(ctx context.Context, actor policy.Actor, filter domain.ProductAliasListFilter) (int, error)
	FindProductByAlias(ctx context.Context, actor policy.Actor, alias string) (domain.ProductDetails, error)
}
type ProductAliasHandler struct {
	aliasService ServiceProductAliasInterface
	validate     *validator.Validate
}

// CreateProductAliasHandler обрабатывает создание нового алиаса для продукта
//
// @Security BearerAuth
// @Summary Create product alias
// @Description Create a new product alias (user's group or specified group for admin)
// @Tags products
// @Accept json
// @Produce json
// @Param request body dto.ProductAliasRequest true "product alias payload"
// @Success 201 {object} dto.ProductAliasResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/products/aliases [post]
func (h ProductAliasHandler) CreateProductAliasHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Декодирование и валидация тела запроса
	var req dto.ProductAliasRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, &req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	// 3. Вызов сервиса для создания алиаса проудукта
	productAlias, err := h.aliasService.Create(ctx, actor, req.ProductID, req.Alias, req.GroupID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"productId": req.ProductID, "alias": req.Alias, "groupId": req.GroupID})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusCreated, dto.ToProductAliasResponse(productAlias))
}

// GetProductAliasHandler возвращает алиас продукта по ID.
//
// @Security BearerAuth
// @Summary Get product alias
// @Description Get product alias by ID
// @Tags products
// @Produce json
// @Param id path int true "alias ID"
// @Success 200 {object} dto.ProductAliasResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/products/aliases/{id} [get]
func (h ProductAliasHandler) GetProductAliasHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Парсинг ID
	aliasID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Вызов сервиса для получения алиаса продукта
	alias, err := h.aliasService.GetByID(ctx, actor, aliasID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"aliasId": aliasID})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToProductAliasResponse(alias))
}

// UpdateProductAliasHandler обновляет алиас продукта по ID.
//
// @Security BearerAuth
// @Summary Update product alias
// @Description Update product alias
// @Tags products
// @Produce json
// @Param id path int true "alias ID"
// @Param request body dto.ProductAliasUpdateRequest true "product alias payload"
// @Success 200 {object} dto.ProductAliasResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/products/aliases/{id} [patch]
func (h ProductAliasHandler) UpdateProductAliasHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Парсинг ID
	aliasID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Декодирование и валидация тела запроса
	var req dto.ProductAliasUpdateRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, &req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	// 4. Вызов сервиса для обновления
	alias, err := h.aliasService.UpdateByID(ctx, actor, aliasID, dto.ToProductAliasUpdateRequest(req))
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"aliasId": aliasID})
		return
	}

	// 5. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToProductAliasResponse(alias))
}

// DeleteProductAliasHandler удаляет алиас продукта по ID.
//
// @Security BearerAuth
// @Summary Delete product alias
// @Description Delete product alias
// @Tags products
// @Produce json
// @Param id path int true "alias ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/products/aliases/{id} [delete]
func (h ProductAliasHandler) DeleteProductAliasHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Парсинг ID
	aliasID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Вызов сервиса для удаления алиаса продукта
	if err := h.aliasService.DeleteByID(ctx, actor, aliasID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"aliasId": aliasID})
		return
	}

	// 4. Успешное удаление без контента
	w.WriteHeader(http.StatusNoContent)
}

// DeleteAllProductAliasesHandler удаляет все алиасы по ID продукта.
//
// @Security BearerAuth
// @Summary Delete all product aliases
// @Description Delete all product aliases
// @Tags products
// @Produce json
// @Param productId query int true "product ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/products/aliases [delete]
func (h ProductAliasHandler) DeleteAllProductAliasesHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Парсинг продукта из запроса
	productIDString := r.URL.Query().Get("productId")
	productID, err := strconv.Atoi(productIDString)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"productId": productIDString})
		return
	}

	// 3. Валидация продукта
	if productID < 1 {
		helpers.WriteDomainError(w, logger, domain.ErrInvalidInput, map[string]any{"productId": productIDString})
		return
	}

	// 3. Вызов сервиса для удаления
	if err := h.aliasService.DeleteAllProductAliases(ctx, actor, productID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"productId": productID})
		return
	}

	// 4. Успешное удаление без контента
	w.WriteHeader(http.StatusNoContent)
}

// ListProductAliasesHandler возвращает список алиасов продукта, доступных пользователю (из его группы).
//
// @Security BearerAuth
// @Summary List product aliases
// @Description List product aliases
// @Tags products
// @Produce json
// @Param group_ids[] query []int false "Group IDs"
// @Param productId query int false "product ID" minimum(1)
// @Param alias query string false "Alias" minimum(1) maximum(50)
// @Param limit query int false "Limit" default(10) minimum(1) maximum(100)
// @Param offset query int false "Offset" default(0) minimum(0)
// @Success 200 {array} dto.ProductAliasResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/products/aliases [get]
func (h ProductAliasHandler) ListProductAliasesHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Биндим query-параметры в структуру
	var queryFilter dto.ProductAliasFilterQuery
	if err := helpers.FormDecoder.Decode(&queryFilter, r.URL.Query()); err != nil {
		logger.WarnContext(ctx, "invalid query parameters: %w", err)
		helpers.WriteError(w, logger, http.StatusBadRequest, "недопустимые параметры запроса")
		return
	}

	// 3. Валидация
	if err := h.validate.Struct(queryFilter); err != nil {
		helpers.WriteDomainError(w, logger, err, queryFilter)
		return
	}

	// 4. Вызов сервиса для получения списка
	aliases, err := h.aliasService.List(ctx, actor, queryFilter.ToDomainFilter())
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"filter": queryFilter})
		return
	}

	// 5. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToProductAliasListResponse(aliases))
}

// CountProductAliasesHandler возвращает количество алиасов продукта.
//
// @Security BearerAuth
// @Summary Count product aliases
// @Description Count product aliases
// @Tags products
// @Produce json
// @Param group_ids[] query []int false "Group IDs"
// @Param productId query int false "product ID" minimum(1)
// @Param alias query string false "Alias" minimum(1) maximum(50)
// @Param limit query int false "Limit" default(10) minimum(1) maximum(100)
// @Param offset query int false "Offset" default(0) minimum(0)
// @Success 200 {object} dto.CountResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/products/aliases/count [get]
func (h ProductAliasHandler) CountProductAliasesHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Биндим query-параметры в структуру
	var queryFilter dto.ProductAliasFilterQuery
	if err := helpers.FormDecoder.Decode(&queryFilter, r.URL.Query()); err != nil {
		logger.WarnContext(ctx, "invalid query parameters: %w", err)
		helpers.WriteError(w, logger, http.StatusBadRequest, "недопустимые параметры запроса")
		return
	}

	// 3. Валидация
	if err := h.validate.Struct(queryFilter); err != nil {
		helpers.WriteDomainError(w, logger, err, queryFilter)
		return
	}

	// 4. Вызов сервиса для получения количества
	count, err := h.aliasService.Count(ctx, actor, queryFilter.ToDomainFilter())
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"filter": queryFilter})
		return
	}

	// 5. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.CountResponse{Count: count})
}

// FindProductByAliasHandler возвращает продукт по алиасу
//
// @Security BearerAuth
// @Summary Find product by alias
// @Description Find product by alias
// @Tags products
// @Accept json
// @Produce json
// @Param alias query string true "alias"
// @Success 200 {object} dto.ProductResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/products/by-alias [get]
func (h ProductAliasHandler) FindProductByAliasHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Парсинг и валидация алиаса из запроса
	alias, err := helpers.ParsePositiveStrQuery(r, "alias")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Вызов сервиса для поиска продукта
	product, err := h.aliasService.FindProductByAlias(ctx, actor, alias)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"alias": alias})
		return
	}

	// 4. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToProductResponse(product))
}
