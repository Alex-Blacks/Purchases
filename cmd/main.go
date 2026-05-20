package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/config"
	"github.com/Alex-Blacks/Purchases/internal/db/storage"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	logger := logging.NewLogger()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	pool, err := pgxpool.New(ctx, cfg.DBurl)
	if err != nil {
		log.Fatalf("db connection error: %v", err)
	}
	defer pool.Close()

	// storage layer
	st := storage.NewStorage(pool)

	// repositories
	userRepo := storage.NewUserRepo()
	orderRepo := storage.NewOrderRepo()
	orderItemRepo := storage.NewOrderItemRepo()
	storeRepo := storage.NewStoreRepo()
	categoryRepo := storage.NewCategoryRepo()
	productRepo := storage.NewProductRepo()
	productAliasRepo := storage.NewProductAliasRepo()

	// service layer
	svc := service.NewService(
		st,
		userRepo,
		orderRepo,
		orderItemRepo,
		storeRepo,
		categoryRepo,
		productRepo,
		productAliasRepo,
	)

	authSvc := service.NewAuthService(svc, cfg.JWTSecret)

	// routers
	publicRouter := handler.PublicRouter(svc, authSvc, logger)
	privateRouter := handler.PrivateRouter(svc, cfg.JWTSecret, logger)

	// mux root
	mux := http.NewServeMux()

	// mount points
	mux.Handle("/api/public/", http.StripPrefix("/api/public", publicRouter))
	mux.Handle("/api/private/", http.StripPrefix("/api/private", privateRouter))

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	log.Println("server started on :8080")
	log.Fatal(server.ListenAndServe())
}
