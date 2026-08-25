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
	List(ctx context.Context, actor policy.Actor) ([]domain.OrderDetails, error)
	ListAll(ctx context.Context, actor policy.Actor) ([]domain.OrderDetails, error)

	AddItem(ctx context.Context, actor policy.Actor, orderID int, productID int, unitID int, quantity int, groupID *int) (domain.OrderItemDetails, error)
	AddListItems(ctx context.Context, actor policy.Actor, orderID int, items []domain.OrderItemCreate, groupID *int) error
	UpdateItem(ctx context.Context, actor policy.Actor, orderID int, productID int, updateOrder domain.OrderItemUpdate) (domain.OrderItemDetails, error)
	UpdateListItems(ctx context.Context, actor policy.Actor, orderID int, items []domain.OrderItemCreate, groupID *int) error
	DeleteItem(ctx context.Context, actor policy.Actor, orderID int, productID int) error
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
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
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

// ListOrdersHandler возвращает список заказов, доступных пользователю (из его группы).
//
// @Security BearerAuth
// @Summary List user's orders
// @Description Get list of orders belonging to user's group
// @Tags orders
// @Produce json
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

	// 2. Вызов сервиса для получения списка
	orders, err := h.orderService.List(ctx, actor)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 3. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToOrderListResponse(orders))
}

// ListAllOrdersHandler возвращает список всех заказов (только для администраторов).
//
// @Security BearerAuth
// @Summary List all orders (admin only)
// @Description Get list of all orders (requires admin role)
// @Tags orders
// @Produce json
// @Success 200 {array} dto.OrderDetailsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/orders/all [get]
func (h OrderHandler) ListAllOrdersHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение данных из контекста
	ctx := r.Context()
	logger := logging.LoggerFromContext(ctx)
	actor, ok := actorctx.ActorFromContext(ctx)
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 2. Вызов сервиса для получения списка
	orders, err := h.orderService.ListAll(ctx, actor)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	// 3. Преобразование и отправка ответа
	helpers.WriteJSON(w, logger, http.StatusOK, dto.ToOrderListResponse(orders))
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
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
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

// AddListItemsHandler godoc
//
// @Security BearerAuth
// @Summary Add order list items
// @Description Add order list items
// @Tags orders
// @Accept json
// @Produce json
// @Param orderId path int true "order ID"
// @Param request body dto.ListItemsRequest true "item payload"
// @Success 201 "Created"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/orders/{orderId}/list_items [post]
func (h OrderHandler) AddListItemsHandler(w http.ResponseWriter, r *http.Request) {
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
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	items := dto.ToItemListRequest(req)

	if err := h.orderService.AddListItems(ctx, actor, orderID, items, req.GroupID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{
			"orderId": orderID,
			"request": req,
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// 	UpdateItem(ctx context.Context, actor policy.Actor, orderID int, productID int, updateOrder domain.OrderItemUpdate) (domain.OrderItemDetails, error)
// 	UpdateListItems(ctx context.Context, actor policy.Actor, orderID int, items []domain.OrderItemCreate, groupID *int) error
// 	DeleteItem(ctx context.Context, actor policy.Actor, orderID int, productID int) error
// 	FindProductInOrders(ctx context.Context, actor policy.Actor, productID int, groupID *int) ([]domain.OrderItemFindDetails, error)

// UpdateListItemsHandler godoc
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
// @Router /private/orders/{orderId}/list_items [put]
func (h OrderHandler) UpdateListItemsHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := actorctx.ActorFromContext(r.Context())
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	orderID, err := helpers.ParsePositiveIntParam(r, "orderId")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	var req dto.ListItemsRequest

	if err := helpers.DecodeJSON(w, r, logger, &req); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	items := dto.ToItemsRequest(req)

	for _, item := range items {
		if item.Quantity <= 0 || item.ProductID <= 0 {
			helpers.WriteError(w, logger, http.StatusBadRequest, "invalid input")
			return
		}
	}

	err = h.orderService.UpdateListItems(r.Context(), actor, orderID, items)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{
			"orderId": orderID,
			"request": req,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
}

// UpdateItemHandler godoc
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
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := actorctx.ActorFromContext(r.Context())
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

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

	var req dto.ItemUpdateRequest

	if err := helpers.DecodeJSON(w, r, logger, &req); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	if err := helpers.ValidatePositiveInt("quantity", req.Quantity); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.orderService.UpdateItem(r.Context(), actor, orderID, productID, req.UnitID, req.Quantity)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{
			"orderId":   orderID,
			"productId": productID,
			"request":   req,
		})
		return
	}

	resp := dto.ItemDetailsResponse{
		ID:       item.ID,
		Title:    item.Title,
		Quantity: item.Quantity,
	}

	helpers.WriteJSON(w, logger, http.StatusOK, resp)
}

// DeleteItemHandler godoc
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
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := actorctx.ActorFromContext(r.Context())
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

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
	if err := h.orderService.DeleteItem(r.Context(), actor, orderID, productID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{
			"orderId":   orderID,
			"productId": productID,
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
