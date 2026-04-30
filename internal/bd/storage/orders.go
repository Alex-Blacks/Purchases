package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type OrderRepo struct {
	storage *Storage
}

func NewOrderRepo(st *Storage) *OrderRepo {
	return &OrderRepo{
		storage: st,
	}
}

func (r *OrderRepo) CreateOrderTx(ctx context.Context, tx pgx.Tx, userID, storeID int) (int, error) {
	var id int
	if err := tx.QueryRow(ctx, "INSERT INTO orders(user_id, store_id) VALUES ($1, $2) RETURNING id", userID, storeID).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return id, domain.ErrAlreadyExists
		}
		return 0, fmt.Errorf("create order failed: %w", err)
	}

	return id, nil
}

func (r *OrderRepo) GetOrderTx(ctx context.Context, tx pgx.Tx, userID, orderID int) ([]domain.OrderWithItemsDTO, error) {

	rows, err := r.storage.pool.Query(ctx, "SELECT orders.id, users.name, stores.name FROM orders JOIN users ON orders.user_id = users.id JOIN stores ON orders.store_id = stores.id")
	if err != nil {
		return nil, fmt.Errorf("Error get orders: %w", err)
	}
	defer rows.Close()

	var order []domain.OrderWithItemsDTO
	for rows.Next() {
		var order domain.OrderDTO
		var items domain.OrderItemsDTO

		if err := rows.Scan(&items.Order.Id)
	}
}
