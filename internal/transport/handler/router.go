package handler

import (
	"log/slog"

	"github.com/Alex-Blacks/Purchases/internal/transport/middleware"
	"github.com/go-chi/chi/v5"
)

func PrivateRouter(h *Handlers, secret string, logger *slog.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RecoveryMiddleware)
	router.Use(middleware.RequestIDMiddleware)
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.AuthMiddleware(secret))

	// Users
	router.Route("/users", func(r chi.Router) {
		r.Get("/", h.User.ListUsersHandler)

		r.Put("/{userId}", h.User.UpdateUserHandler)
		r.Get("/{userId}", h.User.GetUserByIDHandler)
		r.Delete("/{userId}", h.User.DeleteUserHandler)
	})

	// Products
	router.Route("/products", func(r chi.Router) {
		r.Post("/", h.Product.CreateProductHandler)
		r.Get("/", h.Product.ListProductsHandler)

		r.Get("/{productId}", h.Product.GetProductHandler)
		r.Delete("/{productId}", h.Product.DeleteProductHandler)

		// Поиск по алиасу (query param)
		r.Get("/by-alias", h.Product.FindProductByAliasHandler)

		// ProductsAliase
		r.Route("/{productId}/aliases", func(r chi.Router) {
			r.Post("/", h.Product.CreateProductAliasHandler)
			r.Get("/", h.Product.ListProductAliasesHandler)
			r.Delete("/", h.Product.DeleteAllProductAliasesHandler)

			r.Get("/{aliasId}", h.Product.GetProductAliasHandler)
			r.Delete("/{aliasId}", h.Product.DeleteProductAliasHandler)
		})
	})

	// Categories
	router.Route("/categories", func(r chi.Router) {
		r.Post("/", h.Category.CreateCategoryHandler)
		r.Get("/", h.Category.ListCategoriesHandler)

		r.Get("/{categoryId}", h.Category.GetCategoryHandler)
		r.Delete("/{categoryId}", h.Category.DeleteCategoryHandler)
	})

	// Stores
	router.Route("/stores", func(r chi.Router) {
		r.Post("/", h.Store.CreateStoreHandler)
		r.Get("/", h.Store.ListStoresHandler)

		r.Get("/{storeId}", h.Store.GetStoreHandler)
		r.Delete("/{storeId}", h.Store.DeleteStoreHandler)
	})

	// Orders
	router.Route("/orders", func(r chi.Router) {
		r.Post("/", h.Order.CreateOrderHandler)
		r.Get("/", h.Order.ListOrdersHandler)

		r.Get("/{orderId}", h.Order.GetOrderHandler)
		r.Delete("/{orderId}", h.Order.DeleteOrderHandler)

		// Items
		r.Route("/{orderId}/items", func(r chi.Router) {
			r.Post("/", h.Order.AddItemHandler)

			r.Put("/{productId}", h.Order.UpdateItemHandler)
			r.Delete("/{productId}", h.Order.DeleteItemHandler)
		})
	})
	return router
}

func PublicRouter(h *Handlers, logger *slog.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RecoveryMiddleware)
	router.Use(middleware.RequestIDMiddleware)
	router.Use(middleware.LoggingMiddleware(logger))

	//Login
	router.Route("/login", func(r chi.Router) {
		r.Post("/", h.Auth.LoginHandler)
	})

	// Users
	router.Route("/users", func(r chi.Router) {
		r.Post("/", h.User.CreateUserHandler)
	})

	return router
}
