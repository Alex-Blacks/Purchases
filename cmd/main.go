package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alex-Blacks/Purchases/internal/config"
	"github.com/Alex-Blacks/Purchases/internal/db/storage"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	if err := godotenv.Load(); err != nil {
		log.Fatal("error loaded .env file")
	}
	cfg := config.Load()
	logger := logging.NewLogger()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connection error: %v", err)
	}
	defer pool.Close()

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
		Addr:    ":" + cfg.AppPort,
		Handler: mux,
	}

	go func() {
		log.Printf("server started on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			log.Printf("server error: %v", err)
			stop()
		}
	}()

	<-ctx.Done()

	log.Println("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("error shutdown")
	}

	log.Println("server stopped")

}
