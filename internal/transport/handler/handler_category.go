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

type ServiceCategoryInterface interface {
	CreateCategory(ctx context.Context, name string) (int, error)
	GetCategory(ctx context.Context, id int) (domain.Category, error)
	DeleteCategory(ctx context.Context, id int) error
	ListCategories(ctx context.Context) ([]domain.Category, error)
}

type CategoryHandler struct {
	categoryService ServiceCategoryInterface
}

func (h CategoryHandler) CreateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())
	var req dto.CategoryRequest

	if err := helpers.DecodeJSON(w, r, logger, &req); err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		helpers.WriteError(w, logger, http.StatusBadRequest, "empty name")
		return
	}

	categoryID, err := h.categoryService.CreateCategory(r.Context(), req.Name)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, req)
		return
	}

	resp := dto.CategoryCreateResponse{
		CategoryID: categoryID,
	}

	helpers.WriteJSON(w, logger, http.StatusCreated, resp)
}

func (h CategoryHandler) GetCategoryHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	categoryID, err := helpers.ParsePositiveIntParam(r, "categoryId")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	category, err := h.categoryService.GetCategory(r.Context(), categoryID)
	if err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"categoryId": categoryID})
		return
	}

	resp := dto.CategoryResponse{
		ID:   category.ID,
		Name: category.Name,
	}

	helpers.WriteJSON(w, logger, http.StatusOK, resp)
}

func (h CategoryHandler) DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	categoryID, err := helpers.ParsePositiveIntParam(r, "categoryId")
	if err != nil {
		helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.categoryService.DeleteCategory(r.Context(), categoryID); err != nil {
		helpers.WriteDomainError(w, logger, err, map[string]any{"categoryId": categoryID})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h CategoryHandler) ListCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.LoggerFromContext(r.Context())

	categories, err := h.categoryService.ListCategories(r.Context())
	if err != nil {
		helpers.WriteDomainError(w, logger, err, nil)
		return
	}
	resp := dto.ToCategoryResponse(categories)

	helpers.WriteJSON(w, logger, http.StatusOK, resp)
}
