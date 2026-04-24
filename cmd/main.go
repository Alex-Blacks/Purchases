package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://dev:devpass@localhost:5432/purchases?ssh=disabled")
	if err != nil {
		log.Fatalf("[ERROR]: Database connection error: %v", err)
	}
	defer pool.Close()
}
