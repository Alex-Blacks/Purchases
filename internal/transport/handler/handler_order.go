package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/pkg"
)

func CreateOrderHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		userID, ok := getIntParam(w, r, "userId", logger)
		if !ok {
			return
		}

		var req struct {
			StoreID int `json:"storeId"`
		}

		if !decodeHelper(w, r, logger, &req) {
			return
		}

		if !validatePositiveInt(w, "userId", userID, logger) {
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(map[string]int{"id": orderID}); err != nil {
			logger.Error("encoding response failed", "error", err)
		}
	}
}

func GetOrderHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		userID, ok := getIntParam(w, r, "userId", logger)
		if !ok {
			return
		}
		orderID, ok := getIntParam(w, r, "orderId", logger)
		if !ok {
			return
		}

		if !validatePositiveInt(w, "userId", userID, logger) {
			return
		}
		if !validatePositiveInt(w, "orderId", orderID, logger) {
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(ToResponse(order)); err != nil {
			logger.Error("encoding response failed", "error", err)
		}
	}
}

func DeleteOrderHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		userID, ok := getIntParam(w, r, "userId", logger)
		if !ok {
			return
		}
		orderID, ok := getIntParam(w, r, "orderId", logger)
		if !ok {
			return
		}

		if !validatePositiveInt(w, "userId", userID, logger) {
			return
		}
		if !validatePositiveInt(w, "orderId", orderID, logger) {
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

		userID, ok := getIntParam(w, r, "userId", logger)
		if !ok {
			return
		}
		if !validatePositiveInt(w, "userId", userID, logger) {
			return
		}

		orders, err := svc.ListOrders(r.Context(), userID)
		if err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"userId": userID,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(orders); err != nil {
			logger.Error("encoding response failed", "error", err)
		}
	}
}
