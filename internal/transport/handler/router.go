package handler

import (
	"log/slog"

	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/middleware"
	"github.com/go-chi/chi/v5"
)

func PrivateRouter(svc *service.Service, secret string, logger *slog.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RecoveryMiddleware)
	router.Use(middleware.RequestIDMiddleware)
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.AuthMiddleware(secret))

	// Users
	router.Route("/users", func(r chi.Router) {
		r.Get("/", ListUsersHandler(svc))

		r.Put("/{userId}", UpdateUserHandler(svc))
		r.Get("/{userId}", GetUserByIDHandler(svc))
		r.Delete("/{userId}", DeleteUserHandler(svc))
	})

	// Products
	router.Route("/products", func(r chi.Router) {
		r.Post("/", CreateProductHandler(svc))
		r.Get("/", ListProductsHandler(svc))

		r.Get("/{productId}", GetProductHandler(svc))
		r.Delete("/{productId}", DeleteProductHandler(svc))

		// Поиск по алиасу (query param)
		r.Get("/by-alias", FindProductByAliasHandler(svc))

		// ProductsAliase
		r.Route("/{productId}/aliases", func(r chi.Router) {
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
	router.Route("/orders", func(r chi.Router) {
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

func PublicRouter(svc *service.Service, auth *service.AuthService, logger *slog.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RecoveryMiddleware)
	router.Use(middleware.RequestIDMiddleware)
	router.Use(middleware.LoggingMiddleware(logger))

	//Login
	router.Route("/login", func(r chi.Router) {
		r.Post("/", LoginHandler(auth))
	})

	// Users
	router.Route("/users", func(r chi.Router) {
		r.Post("/", CreateUserHandler(svc))
	})

	return router
}
