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

type ServiceOrderInterface interface {
	Create(ctx context.Context, actor policy.Actor, storeID int, groupID *int) (domain.OrderCreateDetails, error)
	GetByID(ctx context.Context, actor policy.Actor, orderID int) (domain.OrderWithItemDetails, error)
	DeleteByID(ctx context.Context, actor policy.Actor, orderID int) error
	List(ctx context.Context, actor policy.Actor, filter domain.OrderListFilter) ([]domain.OrderDetails, error)
	Count(ctx context.Context, actor policy.Actor, filter domain.OrderListFilter) (int, error)

	AddItem(ctx context.Context, actor policy.Actor, orderID int, productID int, unitID int, quantity int, groupID *int) (domain.OrderItemDetails, error)
	UpdateItem(ctx context.Context, actor policy.Actor, orderID int, productID int, updateOrder domain.OrderItemUpdate) (domain.OrderItemDetails, error)
	UpsertListItems(ctx context.Context, actor policy.Actor, orderID int, items []domain.OrderItemCreate, groupID *int) error
	DeleteItem(ctx context.Context, actor policy.Actor, orderID int, productID int) error
	ListItems(ctx context.Context, actor policy.Actor, filter domain.OrderItemListFilter) ([]domain.OrderItemDetails, error)
	CountItems(ctx context.Context, actor policy.Actor, filter domain.OrderItemListFilter) (int, error)
	FindProductInOrders(ctx context.Context, actor policy.Actor, productID int, groupID *int) ([]domain.OrderItemFindDetails, error)
}

type OrderHandler struct {
	orderService ServiceOrderInterface
	validate     *validator.Validate
}

// CreateOrderHandler обрабатывает создание нового заказа.
//
// @Security BearerAuth
// @Summary Create order
// @Description Create a new order (user's group or specified group for admin)
// @Tags orders
// @Accept json
// @Produce json
// @Param request body dto.OrderRequest true "order payload"
// @Success 201 {object} dto.OrderDetailsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/orders [post]
func (h OrderHandler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Декодирование и валидация тела запроса
	var req dto.OrderRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, &req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	// 3. Вызов сервиса для создания заказа
	order, err := h.orderService.Create(ctx, actor, req.StoreID, req.GroupID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"storeId": req.StoreID, "groupId": req.GroupID})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusCreated, dto.ToOrderResponse(order))
}

// GetOrderHandler возвращает заказ по ID.
//
// @Security BearerAuth
// @Summary Get order
// @Description Get order by ID
// @Tags orders
// @Produce json
// @Param id path int true "order ID"
// @Success 200 {object} dto.OrderWithItemDetailsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/orders/{id} [get]
func (h OrderHandler) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Извлечение и парсинг ID из пути
	orderID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Вызов сервиса для получения заказа
	order, err := h.orderService.GetByID(ctx, actor, orderID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"orderId": orderID})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToOrderWithItemResponse(order))
}

// DeleteOrderHandler удаляет заказ по ID.
//
// @Security BearerAuth
// @Summary Delete order
// @Description Delete order by ID
// @Tags orders
// @Produce json
// @Param id path int true "order ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/orders/{id} [delete]
func (h OrderHandler) DeleteOrderHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Парсинг ID
	orderID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Вызов сервиса для удаления
	if err := h.orderService.DeleteByID(ctx, actor, orderID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"orderId": orderID})
		return
	}

	// 4. Успешное удаление без контента
	w.WriteHeader(http.StatusNoContent)
}

// ListOrdersHandler возвращает список заказов.
//
// @Security BearerAuth
// @Summary List user's orders
// @Description Get list of orders belonging to user's group
// @Tags orders
// @Produce json
// @Param group_ids[] query []int false "Group IDs"
// @Param user_id query string false "User ID" minimum(1)
// @Param store_id query string false "Store ID" minimum(1)
// @Param created_from query string false "Created from (RFC3339)"
// @Param created_to query string false "Created to (RFC3339)"
// @Param updated_from query string false "Updated from (RFC3339)"
// @Param updated_to query string false "Updated to (RFC3339)"
// @Param limit query int false "Limit" default(10) minimum(1) maximum(100)
// @Param offset query int false "Offset" default(0) minimum(0)
// @Success 200 {array} dto.OrderDetailsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/orders [get]
func (h OrderHandler) ListOrdersHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Биндим query-параметры в структуру
	var queryFilter dto.OrderFilterQuery
	if err := helpers.FormDecoder.Decode(&queryFilter, r.URL.Query()); err != nil {
		logger.WarnContext(ctx, "invalid query parameters", "error", err)
		helpers.WriteError(w, logger, http.StatusBadRequest, "недопустимые параметры запроса")
		return
	}

	// 3. Валидация
	if err := h.validate.Struct(queryFilter); err != nil {
		helpers.WriteDomainError(w, logger, err, queryFilter)
		return
	}

	// 4. Вызов сервиса для получения списка
	orders, err := h.orderService.List(ctx, actor, queryFilter.ToDomainFilter())
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 3. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToOrderListResponse(orders))
}

// CountOrdersHandler возвращает количество всех заказов.
//
// @Security BearerAuth
// @Summary Count all orders
// @Description Get count of all orders
// @Tags orders
// @Produce json
// @Param group_ids[] query []int false "Group IDs"
// @Param user_id query string false "User ID" minimum(1)
// @Param store_id query string false "Store ID" minimum(1)
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
// @Router /private/orders/count [get]
func (h OrderHandler) CountOrdersHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Биндим query-параметры в структуру
	var queryFilter dto.OrderFilterQuery
	if err := helpers.FormDecoder.Decode(&queryFilter, r.URL.Query()); err != nil {
		logger.WarnContext(ctx, "invalid query parameters", "error", err)
		helpers.WriteError(w, logger, http.StatusBadRequest, "недопустимые параметры запроса")
		return
	}

	// 3. Валидация
	if err := h.validate.Struct(queryFilter); err != nil {
		helpers.WriteDomainError(w, logger, err, queryFilter)
		return
	}

	// 2. Вызов сервиса для получения количества
	count, err := h.orderService.Count(ctx, actor, queryFilter.ToDomainFilter())
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 3. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.CountResponse{Count: count})
}

// AddItemHandler обрабатывает добавление элемента в заказ.
//
// @Security BearerAuth
// @Summary Add order item
// @Description Add order item
// @Tags orders
// @Accept json
// @Produce json
// @Param orderId path int true "order ID"
// @Param request body dto.ItemRequest true "item payload"
// @Success 201 {object} dto.ItemDetailsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/orders/{orderId}/items [post]
func (h OrderHandler) AddItemHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Извлечение и парсинг ID из пути
	orderID, err := helpers.ParsePositiveIntParam(r, "orderId")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Декодирование и валидация тела запроса
	var req dto.ItemRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, &req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	// 4. Вызов сервиса для добавления элемента
	item, err := h.orderService.AddItem(ctx, actor, orderID, req.ProductID, req.UnitID, req.Quantity, req.GroupID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{
			"orderId":  orderID,
			"request":  req,
			"quantity": req.Quantity,
			"groupId":  req.GroupID,
		})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusCreated, dto.ToItemResponse(item))
}

// UpdateItemHandler обрабатывает обновление элементов в заказе.
//
// @Security BearerAuth
// @Summary Update order item
// @Description Update order item
// @Tags orders
// @Accept json
// @Produce json
// @Param orderId path int true "order ID"
// @Param productId path int true "product ID"
// @Param request body dto.ItemUpdateRequest true "item payload"
// @Success 200 {object} dto.ItemDetailsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/orders/{orderId}/items/{productId} [patch]
func (h OrderHandler) UpdateItemHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Извлечение и парсинг ID из пути
	orderID, err := helpers.ParsePositiveIntParam(r, "orderId")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}
	productID, err := helpers.ParsePositiveIntParam(r, "productId")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Декодирование и валидация тела запроса
	var req dto.ItemUpdateRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, &req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	// 4. Вызов сервиса для обновления
	item, err := h.orderService.UpdateItem(ctx, actor, orderID, productID, dto.ToItemUpdateRequest(req))
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{
			"orderId":   orderID,
			"productId": productID,
			"request":   req,
		})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToItemResponse(item))
}

// UpsertListItemsHandler обрабатывает обновление слайса элементов в заказе.
//
// @Security BearerAuth
// @Summary Update order list items
// @Description Update order list items
// @Tags orders
// @Accept json
// @Produce json
// @Param orderId path int true "order ID"
// @Param request body dto.ListItemsRequest true "item payload"
// @Success 200 "OK"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/orders/{orderId}/items/list [put]
func (h OrderHandler) UpsertListItemsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Извлечение и парсинг ID из пути
	orderID, err := helpers.ParsePositiveIntParam(r, "orderId")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Декодирование и валидация тела запроса
	var req dto.ListItemsRequest
	if err := helpers.DecodeJSON(w, r, logger, h.validate, &req); err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	// 4. Вызов сервиса для обновления
	items := dto.ToItemListRequest(req)
	err = h.orderService.UpsertListItems(ctx, actor, orderID, items, req.GroupID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{
			"orderId": orderID,
			"request": req,
		})
		return
	}

	// 4. Отправка ответа
	w.WriteHeader(http.StatusOK)
}

// DeleteItemHandler обрабатывает удаление элементов в заказе.
//
// @Security BearerAuth
// @Summary Delete order item
// @Description Delete order item
// @Tags orders
// @Produce json
// @Param orderId path int true "order ID"
// @Param productId path int true "product ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/orders/{orderId}/items/{productId} [delete]
func (h OrderHandler) DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Извлечение и парсинг ID из пути
	orderID, err := helpers.ParsePositiveIntParam(r, "orderId")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}
	productID, err := helpers.ParsePositiveIntParam(r, "productId")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Вызов сервиса для удаления
	if err := h.orderService.DeleteItem(ctx, actor, orderID, productID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{
			"orderId":   orderID,
			"productId": productID,
		})
		return
	}

	// 4. Успешное удаление без контента
	w.WriteHeader(http.StatusNoContent)
}

// ListOrderItemsHandler возвращает список позиций в заказе.
//
// @Security BearerAuth
// @Summary List order items
// @Description Get list of order items
// @Tags orders
// @Produce json
// @Param group_ids[] query []int false "Group IDs"
// @Param order_id query string false "Order ID" minimum(1)
// @Param product_id query string false "Product ID" minimum(1)
// @Param unit_id query string false "Unit ID" minimum(1)
// @Param quantity_min query int false "Minimum quantity"
// @Param quantity_max query int false "Maximum quantity"
// @Param limit query int false "Limit" default(10) minimum(1) maximum(100)
// @Param offset query int false "Offset" default(0) minimum(0)
// @Success 200 {array} dto.ItemDetailsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/orders/items [get]
func (h OrderHandler) ListOrderItemHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Биндим query-параметры в структуру
	var queryFilter dto.OrderItemFilterQuery
	if err := helpers.FormDecoder.Decode(&queryFilter, r.URL.Query()); err != nil {
		logger.WarnContext(ctx, "invalid query parameters", "error", err)
		helpers.WriteError(w, logger, http.StatusBadRequest, "недопустимые параметры запроса")
		return
	}

	// 3. Валидация
	if err := h.validate.Struct(queryFilter); err != nil {
		helpers.WriteDomainError(w, logger, err, queryFilter)
		return
	}

	// 4. Вызов сервиса для получения списка
	orders, err := h.orderService.ListItems(ctx, actor, queryFilter.ToDomainFilter())
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 3. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToOrderItemListResponse(orders))
}

// CountOrdersHandler возвращает количество всех продуктов в заказе.
//
// @Security BearerAuth
// @Summary Count all order items
// @Description Get count of all order items
// @Tags orders
// @Produce json
// @Param group_ids[] query []int false "Group IDs"
// @Param order_id query string false "Order ID" minimum(1)
// @Param product_id query string false "Product ID" minimum(1)
// @Param unit_id query string false "Unit ID" minimum(1)
// @Param quantity_min query int false "Minimum quantity"
// @Param quantity_max query int false "Maximum quantity"
// @Param limit query int false "Limit" default(10) minimum(1) maximum(100)
// @Param offset query int false "Offset" default(0) minimum(0)
// @Success 200 {object} dto.CountResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/orders/items/count [get]
func (h OrderHandler) CountOrderItemsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Биндим query-параметры в структуру
	var queryFilter dto.OrderItemFilterQuery
	if err := helpers.FormDecoder.Decode(&queryFilter, r.URL.Query()); err != nil {
		logger.WarnContext(ctx, "invalid query parameters", "error", err)
		helpers.WriteError(w, logger, http.StatusBadRequest, "недопустимые параметры запроса")
		return
	}

	// 3. Валидация
	if err := h.validate.Struct(queryFilter); err != nil {
		helpers.WriteDomainError(w, logger, err, queryFilter)
		return
	}

	// 2. Вызов сервиса для получения количества
	count, err := h.orderService.CountItems(ctx, actor, queryFilter.ToDomainFilter())
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 3. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.CountResponse{Count: count})
}

// FindProductInOrdersHandler выводит магазины с искомым продуктом
//
// @Security BearerAuth
// @Summary Find product in orders
// @Description Find product in orders
// @Tags orders
// @Produce json
// @Param groupId query int false "Group ID (optional)"
// @Param productId query int  true "product ID"
// @Success 200 {array} dto.OrderItemFindResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/orders/by-productId [get]
func (h OrderHandler) FindProductInOrdersHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Парсинг данных из query
	productID, err := helpers.ParsePositiveIntQuery(r, "productId")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	groupID, err := helpers.ParseOptionalIntParam(r, "groupId")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Вызов сервиса для поиска товара в заказах
	stores, err := h.orderService.FindProductInOrders(ctx, actor, productID, groupID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{
			"productId": productID,
		})
		return
	}

	// 4. Формирование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToOrderItemFindResponse(stores))
}
