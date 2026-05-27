package handler

import (
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/authctx"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
)

func CreateOrderHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		actor, ok := authctx.ActorFromContext(r.Context())
		if !ok {
			return
		}

		var req dto.OrderRequest

		if !helpers.DecodeJSONHelper(w, r, logger, &req) {
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

func GetOrderHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		actor, ok := authctx.ActorFromContext(r.Context())
		if !ok {
			return
		}
		orderID, ok := helpers.ParsePositiveIntParam(w, r, "orderId", logger)
		if !ok {
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

func DeleteOrderHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		actor, ok := authctx.ActorFromContext(r.Context())
		if !ok {
			return
		}
		orderID, ok := helpers.ParsePositiveIntParam(w, r, "orderId", logger)
		if !ok {
			return
		}

		if err := svc.DeleteOrder(r.Context(), actor, orderID); err != nil {
			helpers.WriteDomainError(w, logger, err, map[string]any{"orderId": orderID})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func ListOrdersHandler(svc *service.Service) http.HandlerFunc {
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
