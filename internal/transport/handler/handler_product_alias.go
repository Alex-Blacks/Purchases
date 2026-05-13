package handler

import (
	"net/http"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/pkg"
)

func CreateProductAliasHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		productID, ok := parsePositiveIntParam(w, r, "productId", logger)
		if !ok {
			return
		}

		var req dto.ProductAliasRequest

		if !decodeHelper(w, r, logger, &req) {
			return
		}

		if strings.TrimSpace(req.Alias) == "" {
			logger.Info("empty name")
			http.Error(w, "empty name", http.StatusBadRequest)
			return
		}

		aliasID, err := svc.CreateProductAlias(r.Context(), productID, req.Alias)
		if err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"productId": productID,
				"alias":     req.Alias,
			})
			return
		}

		resp := dto.ProductAliasCreateResponse{
			AliasID: aliasID,
		}

		encodeHelper(w, logger, http.StatusCreated, resp)
	}
}

func GetProductAliasHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		aliasID, ok := parsePositiveIntParam(w, r, "aliasId", logger)
		if !ok {
			return
		}

		alias, err := svc.GetProductAlias(r.Context(), aliasID)
		if err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"aliasId": aliasID,
			})
			return
		}

		resp := dto.ToProductAliasResponse(alias)

		encodeHelper(w, logger, http.StatusOK, resp)
	}
}

func DeleteProductAliasHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		aliasID, ok := parsePositiveIntParam(w, r, "aliasId", logger)
		if !ok {
			return
		}

		if err := svc.DeleteProductAlias(r.Context(), aliasID); err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"aliasId": aliasID,
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func ListProductAliasesHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		productID, ok := parsePositiveIntParam(w, r, "productId", logger)
		if !ok {
			return
		}

		svc.ListProductAliases(r.Context(), productID)
	}
}
