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
		if err := json.NewEncoder(w).Encode(dto.CategoryResponse{
			CategoryID: categoryID,
		}); err != nil {
			logger.Info("encoding response failed", "error", err)
		}
	}
}
