package handler

import (
	"net/http"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
)

func CreateStoreHandler(svc *service.ServiceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		var req dto.StoreRequest

		if err := helpers.DecodeJSON(w, r, logger, &req); err != nil {
			helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
			return
		}
		if strings.TrimSpace(req.Name) == "" {
			helpers.WriteError(w, logger, http.StatusBadRequest, "empty name")
			return
		}

		storeID, err := svc.CreateStore(r.Context(), req.Name)
		if err != nil {
			helpers.WriteDomainError(w, logger, err, req)
			return
		}

		resp := dto.StoreResponse{
			ID:   storeID,
			Name: req.Name,
		}

		helpers.WriteJSON(w, logger, http.StatusCreated, resp)
	}
}

func GetStoreHandler(svc *service.ServiceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		storeID, err := helpers.ParsePositiveIntParam(r, "storeId")
		if err != nil {
			helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
			return
		}

		store, err := svc.GetStore(r.Context(), storeID)
		if err != nil {
			helpers.WriteDomainError(w, logger, err, map[string]any{"storeId": storeID})
			return
		}
		resp := dto.StoreResponse{
			ID:   store.ID,
			Name: store.Name,
		}

		helpers.WriteJSON(w, logger, http.StatusOK, resp)
	}
}

func DeleteStoreHandler(svc *service.ServiceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		storeID, err := helpers.ParsePositiveIntParam(r, "storeId")
		if err != nil {
			helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
			return
		}

		if err := svc.DeleteStore(r.Context(), storeID); err != nil {
			helpers.WriteDomainError(w, logger, err, map[string]any{"storeId": storeID})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

func ListStoresHandler(svc *service.ServiceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())
		list, err := svc.ListStores(r.Context())
		if err != nil {
			helpers.WriteDomainError(w, logger, err, nil)
			return
		}

		resp := dto.ToStoreResponse(list)

		helpers.WriteJSON(w, logger, http.StatusOK, resp)
	})
}
