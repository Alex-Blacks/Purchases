package handler

import (
	"net/http"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/helpers"
)

func CreateCategoryHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())
		var req dto.CategoryRequest

		if !helpers.DecodeJSONHelper(w, r, logger, &req) {
			return
		}

		if strings.TrimSpace(req.Name) == "" {
			logger.Info("empty name")
			http.Error(w, "empty name", http.StatusBadRequest)
			return
		}

		categoryID, err := svc.CreateCategory(r.Context(), req.Name)
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"name": req.Name,
			})
			return
		}

		resp := dto.CategoryCreateResponse{
			CategoryID: categoryID,
		}

		helpers.RespondJSON(w, http.StatusCreated, resp, logger)
	}
}

func GetCategoryHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())
		categoryID, ok := helpers.ParsePositiveIntParam(w, r, "categoryId", logger)
		if !ok {
			return
		}

		category, err := svc.GetCategory(r.Context(), categoryID)
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"categoryId": categoryID,
			})
			return
		}

		resp := dto.CategoryResponse{
			ID:   category.ID,
			Name: category.Name,
		}

		helpers.RespondJSON(w, http.StatusOK, resp, logger)
	}
}

func DeleteCategoryHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())
		categoryID, ok := helpers.ParsePositiveIntParam(w, r, "categoryId", logger)
		if !ok {
			return
		}

		if err := svc.DeleteCategory(r.Context(), categoryID); err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{
				"categoryId": categoryID,
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func ListCategoriesHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.LoggerFromContext(r.Context())

		categories, err := svc.ListCategories(r.Context())
		if err != nil {
			helpers.DomainErrResponse(w, err, logger, map[string]any{})
			return
		}
		resp := dto.ToCategoryResponse(categories)

		helpers.RespondJSON(w, http.StatusOK, resp, logger)
	}
}
