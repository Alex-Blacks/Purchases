package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/service"
)

func HandlerCreateStore(svc *service.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name string `json:"name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		ctx := r.Context()

		if err := svc.CreateStore(ctx, req.Name); err != nil {
			if errors.Is(err, domain.ErrEmptyName) {
				http.Error(w, domain.ErrEmptyName.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}

func HandlerGetStore(svc *service.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := r.Header.Get("id")
		id, err := strconv.Atoi(req)
		if err != nil {
			http.Error(w, domain.ErrInvalidId.Error(), http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		name, err := svc.GetStoreById(ctx, id)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidId) {
				http.Error(w, domain.ErrInvalidId.Error(), http.StatusBadRequest)
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	})
}
