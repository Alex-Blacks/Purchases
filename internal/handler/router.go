package handler

import (
	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/go-chi/chi/v5"
)

func NewRouter(svc *service.Service) *chi.Mux {
	routOrder := chi.NewRouter()

	routOrder.Post("/orders/", CreateOrderHandler(svc))
	routOrder.Get("/orders/", GetOrderHandler(svc))
	routOrder.Delete("/orders/", DeleteOrderHandler(svc))
	routOrder.Post("/orders/list/", ListOrdersHandler(svc))

	return routOrder
}
