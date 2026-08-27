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
		r.Post("/", h.User.CreateUserHandler)
		r.Get("/", h.User.ListUsersHandler)
		r.Get("/all", h.User.ListAllUsersHandler)

		r.Patch("/{id}", h.User.UpdateUserHandler)
		r.Get("/{id}", h.User.GetUserByIDHandler)
		r.Delete("/{id}", h.User.DeleteUserHandler)
	})

	// Products
	router.Route("/products", func(r chi.Router) {
		r.Post("/", h.Product.CreateProductHandler)
		r.Get("/", h.Product.ListProductsHandler)
		r.Get("/all", h.Product.ListAllProductsHandler)

		r.Patch("/{id}", h.Product.UpdateProductHandler)
		r.Get("/{id}", h.Product.GetProductHandler)
		r.Delete("/{id}", h.Product.DeleteProductHandler)

		// Поиск по алиасу (query param)
		r.Get("/by-alias", h.ProductAlias.FindProductByAliasHandler)
		r.Get("/all/by-alias", h.ProductAlias.FindAllProductByAliasHandler)

		// ProductsAliase
		r.Route("/{productId}/aliases", func(r chi.Router) {
			r.Post("/", h.ProductAlias.CreateProductAliasHandler)
			r.Get("/", h.ProductAlias.ListProductAliasesHandler)
			r.Delete("/", h.ProductAlias.DeleteAllProductAliasesHandler)

			r.Get("/all", h.ProductAlias.ListAllProductAliasesHandler)

			r.Patch("/{id}", h.ProductAlias.UpdateProductAliasHandler)
			r.Get("/{id}", h.ProductAlias.GetProductAliasHandler)
			r.Delete("/{id}", h.ProductAlias.DeleteProductAliasHandler)
		})
	})

	// Stores
	router.Route("/stores", func(r chi.Router) {
		r.Post("/", h.Store.CreateStoreHandler)
		r.Get("/", h.Store.ListStoresHandler)
		r.Get("/all", h.Store.ListAllStoresHandler)

		r.Patch("/{id}", h.Store.UpdateStoreHandler)
		r.Get("/{id}", h.Store.GetStoreHandler)
		r.Delete("/{id}", h.Store.DeleteStoreHandler)
	})

	// Units
	router.Route("/units", func(r chi.Router) {
		r.Post("/", h.Unit.CreateUnitHandler)
		r.Get("/", h.Unit.ListUnitsHandler)
		r.Get("/all", h.Unit.ListAllUnitsHandler)

		r.Patch("/{id}", h.Unit.UpdateUnitHandler)
		r.Get("/{id}", h.Unit.GetUnitHandler)
		r.Delete("/{id}", h.Unit.DeleteUnitHandler)
	})

	// Groups
	router.Route("/groups", func(r chi.Router) {
		r.Post("/", h.Group.CreateGroupHandler)
		r.Get("/", h.Group.ListAllGroupsHandler)

		r.Patch("/{id}", h.Group.UpdateGroupHandler)
		r.Get("/{id}", h.Group.GetGroupHandler)
		r.Delete("/{id}", h.Group.DeleteGroupHandler)
	})

	// Invites
	router.Route("/invites", func(r chi.Router) {
		r.Post("/", h.Invite.CreateInviteHandler)
		r.Get("/", h.Invite.ListInvitesHandler)
		r.Get("/all", h.Invite.ListAllInvitesHandler)

		r.Get("/{id}", h.Invite.GetInviteHandler)
		r.Delete("/{id}", h.Invite.DeleteInviteHandler)

		r.Put("/accept", h.Invite.AcceptInviteHandler)
		r.Put("/reject", h.Invite.RejectInviteHandler)
	})

	// Orders
	router.Route("/orders", func(r chi.Router) {
		r.Post("/", h.Order.CreateOrderHandler)
		r.Get("/", h.Order.ListOrdersHandler)
		r.Get("/all", h.Order.ListAllOrdersHandler)

		r.Get("/{id}", h.Order.GetOrderHandler)
		r.Delete("/{id}", h.Order.DeleteOrderHandler)

		r.Post("/{orderId}/list", h.Order.AddListItemsHandler)
		r.Put("/{orderId}/list", h.Order.UpdateListItemsHandler)

		// Поиск по продукту (query param)
		r.Get("/by-productId", h.Order.FindProductInOrdersHandler)

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

	//Register
	router.Route("/register", func(r chi.Router) {
		r.Post("/", h.Auth.RegisterHandler)
	})

	return router
}
