package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/service"
)

func CreateStoreHandler(svc *service.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var req struct {
			Name string `json:"name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		if err := svc.CreateStore(ctx, req.Name); err != nil {
			if errors.Is(err, domain.ErrEmptyName) {
				http.Error(w, domain.ErrEmptyName.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "created"}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func GetStoreHandler(svc *service.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		string_id := r.URL.Query().Get("id")
		id, err := strconv.Atoi(string_id)
		if err != nil {
			http.Error(w, domain.ErrInvalidInput.Error(), http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		name, err := svc.GetStoreById(ctx, id)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidInput) {
				http.Error(w, domain.ErrInvalidInput.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(map[string]string{"Name Store": name}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	})
}

func DeleteStoreHandler(svc *service.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		string_id := r.URL.Query().Get("id")
		id, err := strconv.Atoi(string_id)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if err := svc.DeleteStore(r.Context(), id); err != nil {
			switch {
			case errors.Is(err, domain.ErrInvalidInput):
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			case errors.Is(err, domain.ErrNotFound):
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

func ListStoreHandler(svc *service.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		list, err := svc.ListStore(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(list); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
