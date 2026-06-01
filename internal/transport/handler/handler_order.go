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

func (h OrderHandler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := authctx.ActorFromContext(r.Context())
	if !ok {
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

func (h OrderHandler) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := authctx.ActorFromContext(r.Context())
	if !ok {
		return
	}
	orderID, err := helpers.ParsePositiveIntParam(r, "orderId")
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

func (h OrderHandler) DeleteOrderHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := authctx.ActorFromContext(r.Context())
	if !ok {
		return
	}
	orderID, err := helpers.ParsePositiveIntParam(r, "orderId")
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

func (h OrderHandler) ListOrdersHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := authctx.ActorFromContext(r.Context())
	if !ok {
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

func (h OrderHandler) AddItemHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := authctx.ActorFromContext(r.Context())
	if !ok {
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

func (h OrderHandler) UpdateItemHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := authctx.ActorFromContext(r.Context())
	if !ok {
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

	helpers.WriteJSON(w, logger, http.StatusCreated, resp)
}

func (h OrderHandler) DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := authctx.ActorFromContext(r.Context())
	if !ok {
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
