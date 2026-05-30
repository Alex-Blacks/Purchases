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

func CreateProductAliasHandler(svc *service.ServiceProductAlias) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		productID, err := helpers.ParsePositiveIntParam(r, "productId")
		if err != nil {
			helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
			return
		}

		var req dto.ProductAliasRequest

		if err := helpers.DecodeJSON(w, r, logger, &req); err != nil {
			return
		}

		if strings.TrimSpace(req.Alias) == "" {
			helpers.WriteError(w, logger, http.StatusBadRequest, "empty name")
			return
		}

		aliasID, err := svc.CreateProductAlias(r.Context(), productID, req.Alias)
		if err != nil {
			helpers.WriteDomainError(w, logger, err, req)
			return
		}

		resp := dto.ProductAliasCreateResponse{
			AliasID: aliasID,
		}

		helpers.WriteJSON(w, logger, http.StatusCreated, resp)
	}
}

func GetProductAliasHandler(svc *service.ServiceProductAlias) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		aliasID, err := helpers.ParsePositiveIntParam(r, "aliasId")
		if err != nil {
			helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
			return
		}

		alias, err := svc.GetProductAlias(r.Context(), aliasID)
		if err != nil {
			helpers.WriteDomainError(w, logger, err, map[string]any{"aliasId": aliasID})
			return
		}

		resp := dto.ToProductAliasResponse(alias)

		helpers.WriteJSON(w, logger, http.StatusOK, resp)
	}
}

func DeleteProductAliasHandler(svc *service.ServiceProductAlias) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		aliasID, err := helpers.ParsePositiveIntParam(r, "aliasId")
		if err != nil {
			helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
			return
		}

		if err := svc.DeleteProductAlias(r.Context(), aliasID); err != nil {
			helpers.WriteDomainError(w, logger, err, map[string]any{"aliasId": aliasID})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func ListProductAliasesHandler(svc *service.ServiceProductAlias) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		productID, err := helpers.ParsePositiveIntParam(r, "productId")
		if err != nil {
			helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
			return
		}
		aliases, err := svc.ListProductAliases(r.Context(), productID)
		if err != nil {
			helpers.WriteDomainError(w, logger, err, map[string]any{"productId": productID})
			return
		}

		resp := dto.ToProductAliasesResponse(aliases)

		helpers.WriteJSON(w, logger, http.StatusOK, resp)
	}
}

func DeleteAllProductAliasesHandler(svc *service.ServiceProductAlias) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		productID, err := helpers.ParsePositiveIntParam(r, "productId")
		if err != nil {
			helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
			return
		}

		if err := svc.DeleteAllProductAliases(r.Context(), productID); err != nil {
			helpers.WriteDomainError(w, logger, err, map[string]any{"productId": productID})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
func FindProductByAliasHandler(svc *service.ServiceProductAlias) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		alias := chi.URLParam(r, "alias")

		if strings.TrimSpace(alias) == "" {
			helpers.WriteError(w, logger, http.StatusBadRequest, "empty name")
			return
		}

		product, err := svc.FindProductByAlias(r.Context(), alias)
		if err != nil {
			helpers.WriteDomainError(w, logger, err, map[string]any{"alias": alias})
			return
		}

		resp := dto.ProductFindResponse{
			Product: product,
		}
		helpers.WriteJSON(w, logger, http.StatusOK, resp)
	}
}
