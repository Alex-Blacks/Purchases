package handler

import (
	"log/slog"
	"time"

	"github.com/Alex-Blacks/Purchases/internal/transport/middleware"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

func PrivateRouter(h *Handlers, secret string, timeout time.Duration, logger *slog.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RecoveryMiddleware)
	router.Use(middleware.TimeoutMiddleware(timeout))
	router.Use(middleware.RequestIDMiddleware)
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.AuthMiddleware(secret))

	router.Get("/swagger/*", httpSwagger.WrapHandler)

	// Users
	router.Route("/users", func(r chi.Router) {
		r.Get("/", h.User.ListUsersHandler)

		r.Patch("/{id}", h.User.UpdateUserHandler)
		r.Get("/{id}", h.User.GetUserByIDHandler)
		r.Delete("/{id}", h.User.DeleteUserHandler)
	})

	// Products
	router.Route("/products", func(r chi.Router) {
		r.Post("/", h.Product.CreateProductHandler)
		r.Get("/", h.Product.ListProductsHandler)

		r.Get("/{id}", h.Product.GetProductHandler)
		r.Delete("/{id}", h.Product.DeleteProductHandler)

		// Поиск по алиасу (query param)
		r.Get("/by-alias", h.Product.FindProductByAliasHandler)

		// ProductsAliase
		r.Route("/{productId}/aliases", func(r chi.Router) {
			r.Post("/", h.Product.CreateProductAliasHandler)
			r.Get("/", h.Product.ListProductAliasesHandler)
			r.Delete("/", h.Product.DeleteAllProductAliasesHandler)

			r.Get("/{id}", h.Product.GetProductAliasHandler)
			r.Delete("/{id}", h.Product.DeleteProductAliasHandler)
		})
	})

	// Categories
	router.Route("/categories", func(r chi.Router) {
		r.Post("/", h.Category.CreateCategoryHandler)
		r.Get("/", h.Category.ListCategoriesHandler)

		r.Get("/{id}", h.Category.GetCategoryHandler)
		r.Delete("/{id}", h.Category.DeleteCategoryHandler)
	})

	// Stores
	router.Route("/stores", func(r chi.Router) {
		r.Post("/", h.Store.CreateStoreHandler)
		r.Get("/", h.Store.ListStoresHandler)

		r.Get("/{id}", h.Store.GetStoreHandler)
		r.Delete("/{id}", h.Store.DeleteStoreHandler)
	})

	// Orders
	router.Route("/orders", func(r chi.Router) {
		r.Post("/", h.Order.CreateOrderHandler)
		r.Get("/", h.Order.ListOrdersHandler)

		r.Get("/{id}", h.Order.GetOrderHandler)
		r.Delete("/{id}", h.Order.DeleteOrderHandler)

		// Items
		r.Route("/{orderId}/items", func(r chi.Router) {
			r.Post("/", h.Order.AddItemHandler)

			r.Patch("/{productId}", h.Order.UpdateItemHandler)
			r.Delete("/{productId}", h.Order.DeleteItemHandler)
		})
	})
	return router
}

func PublicRouter(h *Handlers, timeout time.Duration, logger *slog.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RecoveryMiddleware)
	router.Use(middleware.TimeoutMiddleware(timeout))
	router.Use(middleware.RequestIDMiddleware)
	router.Use(middleware.LoggingMiddleware(logger))

	router.Get("/swagger/*", httpSwagger.WrapHandler)

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
