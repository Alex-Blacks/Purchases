package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/pkg"
)

func CreateStoreHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		var req struct {
			Name string `json:"name"`
		}

		if !decodeHelper(w, r, logger, &req) {
			return
		}
		if strings.TrimSpace(req.Name) == "" {
			logger.Info("empty name")
			http.Error(w, "empty name", http.StatusBadRequest)
			return
		}

		storeID, err := svc.CreateStore(r.Context(), req.Name)
		if err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"name": req.Name,
			})
			return
		}
		type res struct {
			StoreID int    `json:"storeId"`
			Name    string `json:"name"`
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(res{
			StoreID: storeID,
			Name:    req.Name,
		}); err != nil {
			logger.Error("encoding response failed", "error", err)
		}
	}
}

func GetStoreHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		storeID, ok := getIntParam(w, r, "storeId", logger)
		if !ok {
			return
		}
		if !validatePositiveInt(w, "storeId", storeID, logger) {
			return
		}

		store, err := svc.GetStore(r.Context(), storeID)
		if err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"storeId": storeID,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(store); err != nil {
			logger.Error("encoding response failed", "error", err)
		}

	}
}

func DeleteStoreHandler(svc *service.Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		storeID, ok := getIntParam(w, r, "storeId", logger)
		if !ok {
			return
		}
		if !validatePositiveInt(w, "storeId", storeID, logger) {
			return
		}

		if err := svc.DeleteStore(r.Context(), storeID); err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"storeId": storeID,
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

func ListStoresHandler(svc *service.Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())
		list, err := svc.ListStores(r.Context())
		if err != nil {
			domainErrResponse(w, err, logger, map[string]any{})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(list); err != nil {
			logger.Error("encoding response failed", "error", err)
		}
	})
}
