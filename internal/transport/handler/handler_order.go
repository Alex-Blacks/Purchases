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

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		defer r.Body.Close()

		var req struct {
			UserID  int `json:"userId"`
			StoreID int `json:"storeId"`
		}

		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		if err := dec.Decode(&req); err != nil {
			logger.Info("decoding failed", "error", err)
			http.Error(w, "bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		if req.UserID <= 0 || req.StoreID <= 0 {
			logger.Info("invalid input", "userId", req.UserID, "storeId", req.StoreID)
			http.Error(w, "invalid input: IDs must be > 0", http.StatusBadRequest)
			return
		}

		orderID, err := svc.CreateOrder(r.Context(), req.UserID, req.StoreID)
		if err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"userId":  req.UserID,
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

		userID, ok := getIntQuery(w, r, "userId", logger)
		if !ok {
			return
		}
		orderID, ok := getIntQuery(w, r, "orderId", logger)
		if !ok {
			return
		}

		if userID <= 0 || orderID <= 0 {
			logger.Info("invalid input", "userId", userID, "orderId", orderID)
			http.Error(w, "invalid input: IDs must be > 0", http.StatusBadRequest)
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

		userID, ok := getIntQuery(w, r, "userId", logger)
		if !ok {
			return
		}
		orderID, ok := getIntQuery(w, r, "orderId", logger)
		if !ok {
			return
		}

		if userID <= 0 || orderID <= 0 {
			logger.Info("invalid input", "userId", userID, "orderId", orderID)
			http.Error(w, "invalid input: IDs must be > 0", http.StatusBadRequest)
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

		userID, ok := getIntQuery(w, r, "userId", logger)
		if !ok {
			return
		}
		if userID <= 0 {
			logger.Info("invalid input", "userId", userID)
			http.Error(w, "invalid input: ID must be > 0", http.StatusBadRequest)
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
