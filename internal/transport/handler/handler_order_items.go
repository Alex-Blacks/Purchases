package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/pkg"
)

func AddItemHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		defer r.Body.Close()

		var req struct {
			OrderID   int `json:"orderId"`
			ProductID int `json:"productId"`
			Quantity  int `json:"quantity"`
		}

		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		if err := dec.Decode(&req); err != nil {
			logger.Info("decoding failed", "error", err)
			http.Error(w, "bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		if req.OrderID <= 0 || req.ProductID <= 0 || req.Quantity <= 0 {
			logger.Info("invalid input", "orderId", req.OrderID, "productId", req.ProductID, "quantity", req.Quantity)
			http.Error(w, "invalid input: IDs and quantity must be > 0", http.StatusBadRequest)
			return
		}

		if err := svc.AddItem(r.Context(), req.OrderID, req.ProductID, req.Quantity); err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"orderId":   req.OrderID,
				"productId": req.ProductID,
				"quantity":  req.Quantity,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		resp := map[string]string{"status": "ok"}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Error("encoding response failed", "error", err)
		}

	}
}

func UpdateItemHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		defer r.Body.Close()

		var req struct {
			OrderID   int `json:"orderId"`
			ProductID int `json:"productId"`
			Quantity  int `json:"quantity"`
		}

		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		if err := dec.Decode(&req); err != nil {
			logger.Info("decoding failed", "error", err)
			http.Error(w, "bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		if req.OrderID <= 0 || req.ProductID <= 0 || req.Quantity <= 0 {
			logger.Info("invalid input", "orderId", req.OrderID, "productId", req.ProductID, "quantity", req.Quantity)
			http.Error(w, "invalid input: IDs and quantity must be > 0", http.StatusBadRequest)
			return
		}

		if err := svc.UpdateItem(r.Context(), req.OrderID, req.ProductID, req.Quantity); err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"orderId":   req.OrderID,
				"productId": req.ProductID,
				"quantity":  req.Quantity,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := map[string]string{"status": "ok"}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Error("encoding response failed", "error", err)
		}
	}
}

func DeleteItemHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		orderID, ok := getIntQuery(w, r, "orderId", logger)
		if !ok {
			return
		}
		productID, ok := getIntQuery(w, r, "productId", logger)
		if !ok {
			return
		}

		if orderID <= 0 || productID <= 0 {
			logger.Info("invalid input", "orderId", orderID, "productId", productID)
			http.Error(w, "invalid input: IDs must be > 0", http.StatusBadRequest)
			return
		}
		if err := svc.DeleteItem(r.Context(), orderID, productID); err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"orderId":   orderID,
				"productId": productID,
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
