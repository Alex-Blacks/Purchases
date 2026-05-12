package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/pkg"
)

func AddItemHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		orderID, ok := getIntParam(w, r, "orderId", logger)
		if !ok {
			return
		}

		var req dto.ItemRequest

		if !decodeHelper(w, r, logger, &req) {
			return
		}

		if !validatePositiveInt(w, "orderId", orderID, logger) {
			return
		}
		if !validatePositiveInt(w, "productId", req.ProductID, logger) {
			return
		}
		if !validatePositiveInt(w, "quantity", req.Quantity, logger) {
			return
		}

		item, err := svc.AddItem(r.Context(), orderID, req.ProductID, req.Quantity)
		if err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"orderId":   orderID,
				"productId": req.ProductID,
				"quantity":  req.Quantity,
			})
			return
		}
		resp := dto.ItemDetailsResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Title:     item.Title,
			Quantity:  item.Quantity,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Error("encoding response failed", "error", err)
		}

	}
}

func UpdateItemHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		orderID, ok := getIntParam(w, r, "orderId", logger)
		if !ok {
			return
		}
		productID, ok := getIntParam(w, r, "productId", logger)
		if !ok {
			return
		}

		var req dto.ItemUpdateRequest

		if !decodeHelper(w, r, logger, &req) {
			return
		}

		if !validatePositiveInt(w, "orderId", orderID, logger) {
			return
		}
		if !validatePositiveInt(w, "productId", productID, logger) {
			return
		}
		if !validatePositiveInt(w, "quantity", req.Quantity, logger) {
			return
		}

		item, err := svc.UpdateItem(r.Context(), orderID, productID, req.Quantity)
		if err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"orderId":   orderID,
				"productId": productID,
				"quantity":  req.Quantity,
			})
			return
		}

		resp := dto.ItemDetailsResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Title:     item.Title,
			Quantity:  item.Quantity,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Error("encoding response failed", "error", err)
		}
	}
}

func DeleteItemHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		orderID, ok := getIntParam(w, r, "orderId", logger)
		if !ok {
			return
		}
		productID, ok := getIntParam(w, r, "productId", logger)
		if !ok {
			return
		}

		if !validatePositiveInt(w, "orderId", orderID, logger) {
			return
		}
		if !validatePositiveInt(w, "productId", productID, logger) {
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
