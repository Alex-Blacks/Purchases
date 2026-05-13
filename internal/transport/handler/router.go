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

	logger := pkg.NewLogger()

	router.Use(middleware.RequestIDMiddleware)
	router.Use(func(next http.Handler) http.Handler {
		return middleware.LoggingMiddleware(next, logger)
	})

	// Categories
	router.Route("/categories", func(r chi.Router) {
		r.Post("/", CreateCategoryHandler(svc))
		r.Get("/", ListCategoriesHandler(svc))

		r.Get("/{categoryId}", GetCategoryHandler(svc))
		r.Delete("/{categoryId}", DeleteCategoryHandler(svc))
	})

	// Stores
	router.Route("/stores", func(r chi.Router) {
		r.Post("/", CreateStoreHandler(svc))
		r.Get("/", ListStoresHandler(svc))

		r.Get("/{storeId}", GetStoreHandler(svc))
		r.Delete("/{storeId}", DeleteStoreHandler(svc))
	})

	// Orders
	router.Route("/users/{userId}/orders", func(r chi.Router) {
		r.Post("/", CreateOrderHandler(svc))
		r.Get("/", ListOrdersHandler(svc))

		r.Get("/{orderId}", GetOrderHandler(svc))
		r.Delete("/{orderId}", DeleteOrderHandler(svc))

		r.Route("/{orderId}/items", func(r chi.Router) {
			r.Post("/", AddItemHandler(svc))

			r.Put("/{productId}", UpdateItemHandler(svc))
			r.Delete("/{productId}", DeleteItemHandler(svc))
		})
	})
	return router
}
