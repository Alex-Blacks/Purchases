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
	router.Use(middleware.AuthMiddleware)

	// Products
	router.Route("/products", func(r chi.Router) {
		r.Post("/", CreateProductHandler(svc))
		r.Get("/", ListProductsHandler(svc))

		r.Get("/{productId}", GetProductHandler(svc))
		r.Delete("/{productId}", DeleteProductHandler(svc))

		// Поиск по алиасу (query param)
		r.Get("/by-alias", FindProductByAliasHandler(svc))

		// ProductsAliase
		router.Route("/{productId}/aliases", func(r chi.Router) {
			r.Post("/", CreateProductAliasHandler(svc))
			r.Get("/", ListProductAliasesHandler(svc))
			r.Delete("/", DeleteAllProductAliasesHandler(svc))

			r.Get("/{aliasId}", GetProductAliasHandler(svc))
			r.Delete("/{aliasId}", DeleteProductAliasHandler(svc))
		})
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

		// Items
		r.Route("/{orderId}/items", func(r chi.Router) {
			r.Post("/", AddItemHandler(svc))

			r.Put("/{productId}", UpdateItemHandler(svc))
			r.Delete("/{productId}", DeleteItemHandler(svc))
		})
	})
	return router
}
