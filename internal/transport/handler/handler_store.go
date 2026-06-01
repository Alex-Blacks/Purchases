package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
)

type ServiceStoreInterface interface {
	CreateStore(ctx context.Context, name string) (int, error)
	GetStore(ctx context.Context, id int) (domain.Store, error)
	DeleteStore(ctx context.Context, id int) error
	ListStores(ctx context.Context) ([]domain.Store, error)
}

type StoreHandler struct {
	storeService ServiceStoreInterface
}

func (h StoreHandler) CreateStoreHandler(w http.ResponseWriter, r *http.Request) {
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

	storeID, err := h.storeService.CreateStore(r.Context(), req.Name)
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

func (h StoreHandler) GetStoreHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	storeID, err := helpers.ParsePositiveIntParam(r, "storeId")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	store, err := h.storeService.GetStore(r.Context(), storeID)
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

func (h StoreHandler) DeleteStoreHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	storeID, err := helpers.ParsePositiveIntParam(r, "storeId")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.storeService.DeleteStore(r.Context(), storeID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"storeId": storeID})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h StoreHandler) ListStoresHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())
	list, err := h.storeService.ListStores(r.Context())
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	resp := dto.ToStoreResponse(list)

	helpers.WriteJSON(w, logger, http.StatusOK, resp)
}
