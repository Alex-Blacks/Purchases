package handler

import (
	"net/http"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
)

func CreateProductHandler(svc *service.ServiceProduct) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		var req dto.ProductRequest

		if err := helpers.DecodeJSON(w, r, logger, &req); err != nil {
			helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
			return
		}
		if strings.TrimSpace(req.Title) == "" {
			helpers.WriteError(w, logger, http.StatusBadRequest, "empty name")
			return
		}
		if strings.TrimSpace(req.Unit) == "" {
			helpers.WriteError(w, logger, http.StatusBadRequest, "empty name")
			return
		}

		if err := helpers.ValidatePositiveInt("categoryId", req.CategoryID); err != nil {
			helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
			return
		}

		productID, err := svc.CreateProduct(r.Context(), req.Title, req.Unit, req.CategoryID)
		if err != nil {
			helpers.WriteDomainError(w, logger, err, req)
			return
		}

		resp := dto.ProductCreateResponse{
			ProductID: productID,
		}
		helpers.WriteJSON(w, logger, http.StatusCreated, resp)
	}
}

func GetProductHandler(svc *service.ServiceProduct) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		productID, err := helpers.ParsePositiveIntParam(r, "productId")
		if err != nil {
			helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
			return
		}

		product, err := svc.GetProduct(r.Context(), productID)
		if err != nil {
			helpers.WriteDomainError(w, logger, err, map[string]any{"productId": productID})
			return
		}

		resp := dto.ToProductResponse(product)
		helpers.WriteJSON(w, logger, http.StatusOK, resp)

	}
}

func DeleteProductHandler(svc *service.ServiceProduct) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		productID, err := helpers.ParsePositiveIntParam(r, "productId")
		if err != nil {
			helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
			return
		}

		if err := svc.DeleteProduct(r.Context(), productID); err != nil {
			helpers.WriteDomainError(w, logger, err, map[string]any{"productId": productID})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func ListProductsHandler(svc *service.ServiceProduct) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		products, err := svc.ListProducts(r.Context())
		if err != nil {
			helpers.WriteDomainError(w, logger, err, nil)
			return
		}

		resp := dto.ToProductsResponse(products)
		helpers.WriteJSON(w, logger, http.StatusOK, resp)
	}
}
