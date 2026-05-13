package handler

import (
	"net/http"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/pkg"
)

func CreateProductHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		var req dto.ProductRequest

		if !decodeHelper(w, r, logger, &req) {
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

		if !validatePositiveInt(w, "categoryId", req.CategoryID, logger) {
			return
		}

		productID, err := svc.CreateProduct(r.Context(), req.Title, req.Unit, req.CategoryID)
		if err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"title":      req.Title,
				"unit":       req.Unit,
				"categoryId": req.CategoryID,
			})
			return
		}

		resp := dto.ProductCreateResponse{
			ProductID: productID,
		}
		encodeHelper(w, logger, http.StatusCreated, resp)
	}
}

func GetProductHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		productID, ok := parsePositiveIntParam(w, r, "productId", logger)
		if !ok {
			return
		}

		product, err := svc.GetProduct(r.Context(), productID)
		if err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"productId": productID,
			})
			return
		}

		resp := dto.ToProductResponse(product)
		encodeHelper(w, logger, http.StatusOK, resp)

	}
}

func DeleteProductHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		productID, ok := parsePositiveIntParam(w, r, "productId", logger)
		if !ok {
			return
		}

		if err := svc.DeleteProduct(r.Context(), productID); err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"productId": productID,
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func ListProductsHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		products, err := svc.ListProducts(r.Context())
		if err != nil {
			domainErrResponse(w, err, logger, map[string]any{})
			return
		}

		resp := dto.ToProductsResponse(products)
		encodeHelper(w, logger, http.StatusOK, resp)
	}
}
