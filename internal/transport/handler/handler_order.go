package handler

import (
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/pkg"
)

func CreateOrderHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		userID, ok := parsePositiveIntParam(w, r, "userId", logger)
		if !ok {
			return
		}

		var req dto.OrderRequest

		if !decodeHelper(w, r, logger, &req) {
			return
		}

		if !validatePositiveInt(w, "storeId", req.StoreID, logger) {
			return
		}

		orderID, err := svc.CreateOrder(r.Context(), userID, req.StoreID)
		if err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"userId":  userID,
				"storeId": req.StoreID,
			})
			return
		}

		resp := dto.OrderCreateResponse{ID: orderID}

		encodeHelper(w, logger, http.StatusCreated, resp)
	}
}

func GetOrderHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		userID, ok := parsePositiveIntParam(w, r, "userId", logger)
		if !ok {
			return
		}
		orderID, ok := parsePositiveIntParam(w, r, "orderId", logger)
		if !ok {
			return
		}

		order, err := svc.GetOrder(r.Context(), userID, orderID)
		if err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"userId":  userID,
				"orderId": orderID,
			})
			return
		}

		resp := dto.ToResponseOrder(order)

		encodeHelper(w, logger, http.StatusOK, resp)
	}
}

func DeleteOrderHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		userID, ok := parsePositiveIntParam(w, r, "userId", logger)
		if !ok {
			return
		}
		orderID, ok := parsePositiveIntParam(w, r, "orderId", logger)
		if !ok {
			return
		}

		if err := svc.DeleteOrder(r.Context(), userID, orderID); err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"userId":  userID,
				"orderId": orderID,
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func ListOrdersHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		userID, ok := parsePositiveIntParam(w, r, "userId", logger)
		if !ok {
			return
		}

		orders, err := svc.ListOrders(r.Context(), userID)
		if err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"userId": userID,
			})
			return
		}

		resp := dto.ToOrderListResponse(orders)
		encodeHelper(w, logger, http.StatusOK, resp)
	}
}
