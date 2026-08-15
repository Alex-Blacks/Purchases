package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/authctx"
	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
)

type ServiceStoreInterface interface {
	CreateStore(ctx context.Context, actor policy.Actor, name string) (domain.StoreDetails, error)
	GetStore(ctx context.Context, actor policy.Actor, storeID int) (domain.StoreDetails, error)
	UpdateStore(ctx context.Context, actor policy.Actor, storeID int, updateStore domain.StoreUpdate) (domain.StoreDetails, error)
	DeleteStore(ctx context.Context, actor policy.Actor, storeID int) error
	ListStores(ctx context.Context, actor policy.Actor) ([]domain.StoreDetails, error)
}

type StoreHandler struct {
	storeService ServiceStoreInterface
}

// CreateStoreHandler godoc
//
// @Security BearerAuth
// @Summary Create store
// @Description Create store
// @Tags stores
// @Accept json
// @Produce json
// @Param request body dto.StoreRequest true "store payload"
// @Success 201 {object} dto.StoreResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/stores [post]
func (h StoreHandler) CreateStoreHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := authctx.ActorFromContext(r.Context())
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.StoreRequest

	if err := helpers.DecodeJSON(w, r, logger, &req); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		helpers.WriteError(w, logger, http.StatusBadRequest, "empty name")
		return
	}

	store, err := h.storeService.CreateStore(r.Context(), actor, req.Name)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	resp := dto.StoreResponse{
		ID:   store.ID,
		Name: store.Name,
	}

	helpers.WriteJSON(w, logger, http.StatusCreated, resp)
}

// GetStoreHandler godoc
//
// @Security BearerAuth
// @Summary Get store
// @Description Get store
// @Tags stores
// @Produce json
// @Param id path int true "store ID"
// @Success 200 {object} dto.StoreResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/stores/{id} [get]
func (h StoreHandler) GetStoreHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := authctx.ActorFromContext(r.Context())
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	storeID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	store, err := h.storeService.GetStore(r.Context(), actor, storeID)
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

// UpdateStoreHandler godoc
//
// @Security BearerAuth
// @Summary Update store
// @Description Update store
// @Tags stores
// @Produce json
// @Param id path int true "store ID"
// @Success 200 {object} dto.StoreResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/stores/{id} [put]
func (h StoreHandler) UpdateStoreHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := authctx.ActorFromContext(r.Context())
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	storeID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	var req dto.StoreRequest
	if err := helpers.DecodeJSON(w, r, logger, &req); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		helpers.WriteError(w, logger, http.StatusBadRequest, "empty name")
		return
	}

	updateStore := domain.StoreUpdate{Name: &req.Name}

	store, err := h.storeService.UpdateStore(r.Context(), actor, storeID, updateStore)
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

// DeleteStoreHandler godoc
//
// @Security BearerAuth
// @Summary Delete store
// @Description Delete store
// @Tags stores
// @Produce json
// @Param id path int true "store ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/stores/{id} [delete]
func (h StoreHandler) DeleteStoreHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := authctx.ActorFromContext(r.Context())
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	storeID, err := helpers.ParsePositiveIntParam(r, "id")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.storeService.DeleteStore(r.Context(), actor, storeID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"storeId": storeID})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListStoresHandler godoc
//
// @Security BearerAuth
// @Summary List store
// @Description List store
// @Tags stores
// @Produce json
// @Success 200 {array} dto.StoreResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /private/stores [get]
func (h StoreHandler) ListStoresHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	actor, ok := authctx.ActorFromContext(r.Context())
	if !ok {
		helpers.WriteError(w, logger, http.StatusUnauthorized, "unauthorized")
		return
	}

	list, err := h.storeService.ListStores(r.Context(), actor)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}

	resp := dto.ToStoreResponse(list)

	helpers.WriteJSON(w, logger, http.StatusOK, resp)
}
