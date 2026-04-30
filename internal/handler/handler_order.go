package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/service"
)

func CreateOrderHandler(svc *service.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var req struct {
			UserID  int `json:"userid"`
			StoreID int `json:"storeid"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		type res struct {
			Id     int    `json:"id"`
			Status string `json:"status"`
		}
		w.Header().Set("Content-Type", "application/json")

		orderID, err := svc.CreateOrder(r.Context(), req.UserID, req.StoreID)
		if err != nil {
			if errors.Is(err, domain.ErrAlreadyExists) {
				if err := json.NewEncoder(w).Encode(&res{
					Id:     orderID,
					Status: domain.ErrAlreadyExists.Error(),
				}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(&res{
			Id:     orderID,
			Status: "created",
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
