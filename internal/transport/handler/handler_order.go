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

type ServiceOrderInterface interface {
	CreateOrder(ctx context.Context, actor policy.Actor, storeID int) (int, error)
	GetOrder(ctx context.Context, actor policy.Actor, orderID int) (domain.OrderWithItemDetails, error)
	DeleteOrder(ctx context.Context, actor policy.Actor, orderID int) error
	ListOrders(ctx context.Context, actor policy.Actor) ([]domain.OrderDetails, error)

	AddItem(ctx context.Context, actor policy.Actor, orderID int, productID int, quantity int) (domain.OrderItemDetails, error)
	UpdateItem(ctx context.Context, actor policy.Actor, orderID int, productID int, quantity int) (domain.OrderItemDetails, error)
	DeleteItem(ctx context.Context, actor policy.Actor, orderID int, productID int) error

	GetAccessibleOrder(ctx context.Context, actor policy.Actor, orderID int) (domain.OrderWithItemDetails, error)
}

type OrderHandler struct {
	orderService ServiceOrderInterface
}

// CreateOrderHandler godoc
//
// @Security BearerAuth
// @Summary Create order
// @Description Create order
// @Tags orders
// @Accept json
// @Produce json
// @Param request body dto.OrderRequest true "order payload"
// @Success 201 {object} dto.OrderCreateResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/orders [post]
func (h OrderHandler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := authctx.ActorFromContext(r.Context())
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.OrderRequest

	if err := helpers.DecodeJSON(w, r, logger, &req); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	orderID, err := h.orderService.CreateOrder(r.Context(), actor, req.StoreID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	resp := dto.OrderCreateResponse{ID: orderID}

	helpers.WriteJSON(w, logger, http.StatusCreated, resp)
}

// GetOrderHandler godoc
//
// @Security BearerAuth
// @Summary Get order
// @Description Get order
// @Tags orders
// @Produce json
// @Param id path int true "order ID"
// @Success 200 {object} dto.OrderWithItemDetailsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/orders/{id} [get]
func (h OrderHandler) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := authctx.ActorFromContext(r.Context())
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}
	orderID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	order, err := h.orderService.GetOrder(r.Context(), actor, orderID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"orderId": orderID})
		return
	}

	resp := dto.ToResponseOrder(order)

	helpers.WriteJSON(w, logger, http.StatusOK, resp)
}

// DeleteOrderHandler godoc
//
// @Security BearerAuth
// @Summary Delete order
// @Description Delete order
// @Tags orders
// @Produce json
// @Param id path int true "order ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/orders/{id} [delete]
func (h OrderHandler) DeleteOrderHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := authctx.ActorFromContext(r.Context())
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}
	orderID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.orderService.DeleteOrder(r.Context(), actor, orderID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"orderId": orderID})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListOrdersHandler godoc
//
// @Security BearerAuth
// @Summary List orders
// @Description List orders
// @Tags orders
// @Produce json
// @Success 200 {array} dto.OrderDetailsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/orders [get]
func (h OrderHandler) ListOrdersHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := authctx.ActorFromContext(r.Context())
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	orders, err := h.orderService.ListOrders(r.Context(), actor)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	resp := dto.ToOrderListResponse(orders)
	helpers.WriteJSON(w, logger, http.StatusOK, resp)
}

// AddItemHandler godoc
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
// @Router /private/orders/{orderId}/items [post]
func (h OrderHandler) AddItemHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := authctx.ActorFromContext(r.Context())
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	orderID, err := helpers.ParsePositiveIntParam(r, "orderId")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	var req dto.ItemRequest

	if err := helpers.DecodeJSON(w, r, logger, &req); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	if err := helpers.ValidatePositiveInt("productId", req.ProductID); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}
	if err := helpers.ValidatePositiveInt("quantity", req.Quantity); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.orderService.AddItem(r.Context(), actor, orderID, req.ProductID, req.Quantity)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{
			"orderId": orderID,
			"request": req,
		})
		return
	}
	resp := dto.ItemDetailsResponse{
		ID:        item.ID,
		ProductID: item.ProductID,
		Title:     item.Title,
		Quantity:  item.Quantity,
	}

	helpers.WriteJSON(w, logger, http.StatusCreated, resp)
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
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/orders/{orderId}/items/{productId} [patch]
func (h OrderHandler) UpdateItemHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := authctx.ActorFromContext(r.Context())
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

	item, err := h.orderService.UpdateItem(r.Context(), actor, orderID, productID, req.Quantity)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{
			"orderId":   orderID,
			"productId": productID,
			"request":   req,
		})
		return
	}

	resp := dto.ItemDetailsResponse{
		ID:        item.ID,
		ProductID: item.ProductID,
		Title:     item.Title,
		Quantity:  item.Quantity,
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
// @Failure 500 {object} dto.ErrorResponse
// @Router /private/orders/{orderId}/items/{productId} [delete]
func (h OrderHandler) DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := authctx.ActorFromContext(r.Context())
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
