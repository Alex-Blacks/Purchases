package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/bd/storage"
	"github.com/Alex-Blacks/Purchases/internal/handler"
	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://dev:devpass@localhost:5432/purchases?sslmode=disable")
	if err != nil {
		log.Fatalf("[ERROR]: Database connection error: %v", err)
	}
	defer pool.Close()

	st := storage.NewStorage(pool)
	orderRepo := storage.NewOrderRepo()
	orderItemRepo := storage.NewOrderItemRepo()
	storeRepo := storage.NewStoreRepo(st)
	svc := service.NewService(st, orderRepo, orderItemRepo, storeRepo)

	mux := handler.NewRouter(svc)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("server started on :8080")

	if err = server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
