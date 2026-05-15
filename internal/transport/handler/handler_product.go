package handler

import (
	"net/http"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
)

func CreateProductHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		var req dto.ProductRequest

		if !helpers.DecodeJSONHelper(w, r, logger, &req) {
			return
		}
		if strings.TrimSpace(req.Title) == "" {
			logger.Info("empty name")
			http.Error(w, "empty name", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.Unit) == "" {
			logger.Info("empty name")
			http.Error(w, "empty name", http.StatusBadRequest)
			return
		}

		if !helpers.ValidatePositiveInt(w, "categoryId", req.CategoryID, logger) {
			return
		}

		productID, err := svc.CreateProduct(r.Context(), req.Title, req.Unit, req.CategoryID)
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"title":      req.Title,
				"unit":       req.Unit,
				"categoryId": req.CategoryID,
			})
			return
		}

		resp := dto.ProductCreateResponse{
			ProductID: productID,
		}
		helpers.RespondJSON(w, http.StatusCreated, resp, logger)
	}
}

func GetProductHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		productID, ok := helpers.ParsePositiveIntParam(w, r, "productId", logger)
		if !ok {
			return
		}

		product, err := svc.GetProduct(r.Context(), productID)
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"productId": productID,
			})
			return
		}

		resp := dto.ToProductResponse(product)
		helpers.RespondJSON(w, http.StatusOK, resp, logger)

	}
}

func DeleteProductHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		productID, ok := helpers.ParsePositiveIntParam(w, r, "productId", logger)
		if !ok {
			return
		}

		if err := svc.DeleteProduct(r.Context(), productID); err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"productId": productID,
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func ListProductsHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		products, err := svc.ListProducts(r.Context())
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{})
			return
		}

		resp := dto.ToProductsResponse(products)
		helpers.RespondJSON(w, http.StatusOK, resp, logger)
	}
}
