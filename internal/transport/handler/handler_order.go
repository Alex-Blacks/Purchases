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

		userID, ok := authctx.UserIDFromContext(r.Context())
		if !ok {
			return
		}

		var req dto.OrderRequest

		if !helpers.DecodeJSONHelper(w, r, logger, &req) {
			return
		}

		orderID, err := svc.CreateOrder(r.Context(), userID, req.StoreID)
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"userId":  userID,
				"storeId": req.StoreID,
			})
			return
		}

		resp := dto.OrderCreateResponse{ID: orderID}

		helpers.RespondJSON(w, http.StatusCreated, resp, logger)
	}
}

func GetOrderHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		userID, ok := authctx.UserIDFromContext(r.Context())
		if !ok {
			return
		}
		orderID, ok := helpers.ParsePositiveIntParam(w, r, "orderId", logger)
		if !ok {
			return
		}

		order, err := svc.GetOrder(r.Context(), userID, orderID)
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"userId":  userID,
				"orderId": orderID,
			})
			return
		}

		resp := dto.ToResponseOrder(order)

		helpers.RespondJSON(w, http.StatusOK, resp, logger)
	}
}

func DeleteOrderHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		userID, ok := authctx.UserIDFromContext(r.Context())
		if !ok {
			return
		}
		orderID, ok := helpers.ParsePositiveIntParam(w, r, "orderId", logger)
		if !ok {
			return
		}

		if err := svc.DeleteOrder(r.Context(), userID, orderID); err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"orderId": orderID,
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func ListOrdersHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		userID, ok := authctx.UserIDFromContext(r.Context())
		if !ok {
			return
		}

		orders, err := svc.ListOrders(r.Context(), userID)
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"userId": userID,
			})
			return
		}

		resp := dto.ToOrderListResponse(orders)
		helpers.RespondJSON(w, http.StatusOK, resp, logger)
	}
}
