package handler

import (
	"net/http"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
	"github.com/go-chi/chi/v5"
)

func CreateProductAliasHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		productID, ok := helpers.ParsePositiveIntParam(w, r, "productId", logger)
		if !ok {
			return
		}

		var req dto.ProductAliasRequest

		if !helpers.DecodeJSONHelper(w, r, logger, &req) {
			return
		}

		if strings.TrimSpace(req.Alias) == "" {
			logger.Info("empty name")
			http.Error(w, "empty name", http.StatusBadRequest)
			return
		}

		aliasID, err := svc.CreateProductAlias(r.Context(), productID, req.Alias)
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"productId": productID,
				"alias":     req.Alias,
			})
			return
		}

		resp := dto.ProductAliasCreateResponse{
			AliasID: aliasID,
		}

		helpers.RespondJSON(w, http.StatusCreated, resp, logger)
	}
}

func GetProductAliasHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		aliasID, ok := helpers.ParsePositiveIntParam(w, r, "aliasId", logger)
		if !ok {
			return
		}

		alias, err := svc.GetProductAlias(r.Context(), aliasID)
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"aliasId": aliasID,
			})
			return
		}

		resp := dto.ToProductAliasResponse(alias)

		helpers.RespondJSON(w, http.StatusOK, resp, logger)
	}
}

func DeleteProductAliasHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		aliasID, ok := helpers.ParsePositiveIntParam(w, r, "aliasId", logger)
		if !ok {
			return
		}

		if err := svc.DeleteProductAlias(r.Context(), aliasID); err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"aliasId": aliasID,
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func ListProductAliasesHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		productID, ok := helpers.ParsePositiveIntParam(w, r, "productId", logger)
		if !ok {
			return
		}

		aliases, err := svc.ListProductAliases(r.Context(), productID)
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{"productId": productID})
			return
		}

		resp := dto.ToProductAliasesResponse(aliases)

		helpers.RespondJSON(w, http.StatusOK, resp, logger)
	}
}

func DeleteAllProductAliasesHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		productID, ok := helpers.ParsePositiveIntParam(w, r, "productId", logger)
		if !ok {
			return
		}

		if err := svc.DeleteAllProductAliases(r.Context(), productID); err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{"productId": productID})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
func FindProductByAliasHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		alias := chi.URLParam(r, "alias")

		if strings.TrimSpace(alias) == "" {
			logger.Info("empty name")
			http.Error(w, "empty name", http.StatusBadRequest)
			return
		}

		productID, err := svc.FindProductByAlias(r.Context(), alias)
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{"alias": alias})
			return
		}

		resp := dto.ProductFindResponse{
			ProductID: productID,
		}
		helpers.RespondJSON(w, http.StatusOK, resp, logger)
	}
}
