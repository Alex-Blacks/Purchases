package handler

import (
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/authctx"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
)

func CreateOrderHandler(svc *service.ServiceOrderItem) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		orderID, err := svc.CreateOrder(r.Context(), actor, req.StoreID)
		if err != nil {
			helpers.WriteDomainError(w, logger, err, req)
			return
		}

		resp := dto.OrderCreateResponse{ID: orderID}

		helpers.WriteJSON(w, logger, http.StatusCreated, resp)
	}
}

func GetOrderHandler(svc *service.ServiceOrderItem) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		order, err := svc.GetOrder(r.Context(), actor, orderID)
		if err != nil {
			helpers.WriteDomainError(w, logger, err, map[string]any{"orderId": orderID})
			return
		}

		resp := dto.ToResponseOrder(order)

		helpers.WriteJSON(w, logger, http.StatusOK, resp)
	}
}

func DeleteOrderHandler(svc *service.ServiceOrderItem) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		if err := svc.DeleteOrder(r.Context(), actor, orderID); err != nil {
			helpers.WriteDomainError(w, logger, err, map[string]any{"orderId": orderID})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func ListOrdersHandler(svc *service.ServiceOrderItem) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		actor, ok := authctx.ActorFromContext(r.Context())
		if !ok {
			return
		}

		orders, err := svc.ListOrders(r.Context(), actor)
		if err != nil {
			helpers.WriteDomainError(w, logger, err, nil)
			return
		}

		resp := dto.ToOrderListResponse(orders)
		helpers.WriteJSON(w, logger, http.StatusOK, resp)
	}
}

func AddItemHandler(svc *service.ServiceOrderItem) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		item, err := svc.AddItem(r.Context(), actor, orderID, req.ProductID, req.Quantity)
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
}

func UpdateItemHandler(svc *service.ServiceOrderItem) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		item, err := svc.UpdateItem(r.Context(), actor, orderID, productID, req.Quantity)
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
}

func DeleteItemHandler(svc *service.ServiceOrderItem) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		if err := svc.DeleteItem(r.Context(), actor, orderID, productID); err != nil {
			helpers.WriteDomainError(w, logger, err, map[string]any{
				"orderId":   orderID,
				"productId": productID,
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
