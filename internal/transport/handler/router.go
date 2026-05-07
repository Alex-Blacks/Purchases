package handler

import (
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/middleware"
	"github.com/Alex-Blacks/Purchases/pkg"
	"github.com/go-chi/chi/v5"
)

func NewRouter(svc *service.Service) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestIDMiddleware)
	router.Use(func(next http.Handler) http.Handler {
		return middleware.LoggingMiddleware(next, pkg.NewLogger())
	})

	router.Post("/stores", CreateStoreHandler(svc))
	router.Get("/stores", GetStoreHandler(svc))
	router.Delete("/stores", DeleteStoreHandler(svc))
	router.Get("/stores/list", ListStoresHandler(svc))

	router.Post("/orders", CreateOrderHandler(svc))
	router.Get("/orders", GetOrderHandler(svc))
	router.Delete("/orders", DeleteOrderHandler(svc))
	router.Get("/orders/list", ListOrdersHandler(svc))

	return router
}
