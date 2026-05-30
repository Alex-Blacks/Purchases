package handler

import (
	"net/http"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
)

func CreateCategoryHandler(svc *service.ServiceCategory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		categoryID, err := svc.CreateCategory(r.Context(), req.Name)
		if err != nil {
			helpers.WriteDomainError(w, logger, err, req)
			return
		}

		resp := dto.CategoryCreateResponse{
			CategoryID: categoryID,
		}

		helpers.WriteJSON(w, logger, http.StatusCreated, resp)
	}
}

func GetCategoryHandler(svc *service.ServiceCategory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		categoryID, err := helpers.ParsePositiveIntParam(r, "categoryId")
		if err != nil {
			helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
			return
		}

		category, err := svc.GetCategory(r.Context(), categoryID)
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
}

func DeleteCategoryHandler(svc *service.ServiceCategory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		categoryID, err := helpers.ParsePositiveIntParam(r, "categoryId")
		if err != nil {
			helpers.WriteError(w, logger, http.StatusBadRequest, err.Error())
			return
		}

		if err := svc.DeleteCategory(r.Context(), categoryID); err != nil {
			helpers.WriteDomainError(w, logger, err, map[string]any{"categoryId": categoryID})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func ListCategoriesHandler(svc *service.ServiceCategory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		categories, err := svc.ListCategories(r.Context())
		if err != nil {
			helpers.WriteDomainError(w, logger, err, nil)
			return
		}
		resp := dto.ToCategoryResponse(categories)

		helpers.WriteJSON(w, logger, http.StatusOK, resp)
	}
}
