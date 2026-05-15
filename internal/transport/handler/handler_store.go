package handler

import (
	"net/http"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
)

func CreateStoreHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		var req dto.StoreRequest

		if !helpers.DecodeJSONHelper(w, r, logger, &req) {
			return
		}
		if strings.TrimSpace(req.Name) == "" {
			logger.Info("empty name")
			http.Error(w, "empty name", http.StatusBadRequest)
			return
		}

		storeID, err := svc.CreateStore(r.Context(), req.Name)
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"name": req.Name,
			})
			return
		}

		resp := dto.StoreResponse{
			ID:   storeID,
			Name: req.Name,
		}

		helpers.RespondJSON(w, http.StatusCreated, resp, logger)
	}
}

func GetStoreHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		storeID, ok := helpers.ParsePositiveIntParam(w, r, "storeId", logger)
		if !ok {
			return
		}

		store, err := svc.GetStore(r.Context(), storeID)
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"storeId": storeID,
			})
			return
		}
		resp := dto.StoreResponse{
			ID:   store.ID,
			Name: store.Name,
		}

		helpers.RespondJSON(w, http.StatusOK, resp, logger)
	}
}

func DeleteStoreHandler(svc *service.Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		storeID, ok := helpers.ParsePositiveIntParam(w, r, "storeId", logger)
		if !ok {
			return
		}

		if err := svc.DeleteStore(r.Context(), storeID); err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"storeId": storeID,
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

func ListStoresHandler(svc *service.Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())
		list, err := svc.ListStores(r.Context())
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{})
			return
		}

		resp := dto.ToStoreResponse(list)

		helpers.RespondJSON(w, http.StatusOK, resp, logger)
	})
}
