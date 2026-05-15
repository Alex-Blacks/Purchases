package handler

import (
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
)

func AddItemHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		orderID, ok := helpers.ParsePositiveIntParam(w, r, "orderId", logger)
		if !ok {
			return
		}

		var req dto.ItemRequest

		if !helpers.DecodeJSONHelper(w, r, logger, &req) {
			return
		}

		if !helpers.ValidatePositiveInt(w, "productId", req.ProductID, logger) {
			return
		}
		if !helpers.ValidatePositiveInt(w, "quantity", req.Quantity, logger) {
			return
		}

		item, err := svc.AddItem(r.Context(), orderID, req.ProductID, req.Quantity)
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
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

		helpers.RespondJSON(w, http.StatusCreated, resp, logger)
	}
}

func UpdateItemHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		orderID, ok := helpers.ParsePositiveIntParam(w, r, "orderId", logger)
		if !ok {
			return
		}
		productID, ok := helpers.ParsePositiveIntParam(w, r, "productId", logger)
		if !ok {
			return
		}

		var req dto.ItemUpdateRequest

		if !helpers.DecodeJSONHelper(w, r, logger, &req) {
			return
		}

		if !helpers.ValidatePositiveInt(w, "quantity", req.Quantity, logger) {
			return
		}

		item, err := svc.UpdateItem(r.Context(), orderID, productID, req.Quantity)
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
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

		helpers.RespondJSON(w, http.StatusOK, resp, logger)
	}
}

func DeleteItemHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		orderID, ok := helpers.ParsePositiveIntParam(w, r, "orderId", logger)
		if !ok {
			return
		}
		productID, ok := helpers.ParsePositiveIntParam(w, r, "productId", logger)
		if !ok {
			return
		}

		if err := svc.DeleteItem(r.Context(), orderID, productID); err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"orderId":   orderID,
				"productId": productID,
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
