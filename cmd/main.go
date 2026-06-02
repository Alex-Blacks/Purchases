package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/Alex-Blacks/Purchases/docs"
	"github.com/Alex-Blacks/Purchases/internal/config"
	"github.com/Alex-Blacks/Purchases/internal/db/storage"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/Alex-Blacks/Purchases/internal/transport/handler"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// @title           Purchases API
// @version         1.0
// @description     API for managing purchases.
// @host            localhost:8080
// @BasePath        /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите токен в формате "Bearer <token>"
func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	_ = godotenv.Load()

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

	// services
	userSvc := service.NewServiceUser(st, userRepo)
	orderSvc := service.NewServiceOrderItem(st, orderRepo, orderItemRepo)
	storeSvc := service.NewServiceStore(st, storeRepo)
	categorySvc := service.NewServiceCategory(st, categoryRepo)
	productSvc := service.NewServiceProduct(st, productRepo)

	authSvc := service.NewAuthService(userSvc, cfg.JWTSecret)

	// handlers
	handlers := handler.NewHandlers(userSvc, storeSvc, productSvc, orderSvc, categorySvc, authSvc)

	// routers
	publicRouter := handler.PublicRouter(handlers, logger)
	privateRouter := handler.PrivateRouter(handlers, cfg.JWTSecret, logger)

	mux := chi.NewRouter()
	mux.Mount("/api", publicRouter)
	mux.Mount("/api/private", privateRouter)

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
