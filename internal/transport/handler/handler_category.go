package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
	"github.com/Alex-Blacks/Purchases/pkg"
)

func CreateCategoryHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())
		var req dto.CategoryRequest

		if !decodeHelper(w, r, logger, &req) {
			return
		}

		if strings.TrimSpace(req.Name) == "" {
			logger.Info("empty name")
			http.Error(w, "empty name", http.StatusBadRequest)
			return
		}

		categoryID, err := svc.CreateCategory(r.Context(), req.Name)
		if err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"name": req.Name,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(dto.CategoryCreateResponse{
			CategoryID: categoryID,
		}); err != nil {
			logger.Info("encoding response failed", "error", err)
		}
	}
}

func GetCategoryHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())
		categoryID, ok := getIntParam(w, r, "categoryId", logger)
		if !ok {
			return
		}

		if !validatePositiveInt(w, "categoryId", categoryID, logger) {
			return
		}

		category, err := svc.GetCategory(r.Context(), categoryID)
		if err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"categoryId": categoryID,
			})
			return
		}

		resp := dto.CategoryResponse{
			ID:   category.ID,
			Name: category.Name,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Info("encoding response failed", "error", err)
		}

	}
}

func DeleteCategoryHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())
		categoryID, ok := getIntParam(w, r, "categoryId", logger)
		if !ok {
			return
		}

		if !validatePositiveInt(w, "categoryId", categoryID, logger) {
			return
		}

		if err := svc.DeleteCategory(r.Context(), categoryID); err != nil {
			domainErrResponse(w, err, logger, map[string]any{
				"categoryId": categoryID,
			})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func ListCategoriesHandler(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := pkg.LoggerFromContext(r.Context())

		categories, err := svc.ListCategories(r.Context())
		if err != nil {
			domainErrResponse(w, err, logger, map[string]any{})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(dto.ToCategoryResponse(categories)); err != nil {
			logger.Info("encoding response failed", "error", err)
		}

	}
}
