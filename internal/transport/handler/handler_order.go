package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/service"
)

func CreateOrderHandler(svc *service.Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var req struct {
			UserID  int `json:"userid"`
			StoreID int `json:"storeid"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if dec.More() {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if err := dec.Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		type res struct {
			ID int `json:"id"`
		}

		orderID, err := svc.CreateOrder(r.Context(), req.UserID, req.StoreID)
		if err != nil {
			if errors.Is(err, domain.ErrAlreadyExists) {
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(res{
					ID: orderID,
				})
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(&res{
			ID: orderID,
		}); err != nil {
			http.Error(w, "error encoding", http.StatusInternalServerError)
			return
		}
	})
}

func GetOrderHandler(svc *service.Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		UserID, err := parseIntQuery(r, "userid")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		OrderID, err := parseIntQuery(r, "orderid")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		order, err := svc.GetOrder(r.Context(), UserID, OrderID)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrNotFound):
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			case errors.Is(err, domain.ErrInvalidInput):
				http.Error(w, "invalid input", http.StatusBadRequest)
				return
			default:
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(ToResponse(order)); err != nil {
			http.Error(w, "error encoding", http.StatusInternalServerError)
			return
		}
	})
}

func DeleteOrderHandler(svc *service.Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		UserID, err := parseIntQuery(r, "userid")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		OrderID, err := parseIntQuery(r, "orderid")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := svc.DeleteOrder(r.Context(), UserID, OrderID); err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

func ListOrdersHandler(svc *service.Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		UserID, err := parseIntQuery(r, "userid")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		order, err := svc.ListOrders(r.Context(), UserID)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(order); err != nil {
			http.Error(w, "error encoding", http.StatusInternalServerError)
			return
		}
	})
}
