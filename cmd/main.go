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
	storeRepo := storage.NewStoreRepo(st)
	svc := service.NewService(st, orderRepo, storeRepo)
	mux := http.NewServeMux()

	mux.Handle("/stores/delete", handler.DeleteStoreHandler(svc))
	mux.Handle("/stores/get", handler.GetStoreHandler(svc))
	mux.Handle("/stores", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handler.CreateStoreHandler(svc).ServeHTTP(w, r)
		case http.MethodGet:
			handler.ListStoreHandler(svc).ServeHTTP(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.Handle("/orders", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handler.CreateOrderHandler(svc).ServeHTTP(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("server started on :8080")

	if err = server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
