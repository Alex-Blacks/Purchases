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

func (r *OrderRepo) GetOrderTx(ctx context.Context, tx pgx.Tx, userID, orderID int) (domain.OrderWithItemsDTO, error) {
	var result domain.OrderWithItemsDTO

	rowsOrder := tx.QueryRow(ctx, `
		SELECT orders.id, users.name, stores.name, orders.created_at, orders.updated_at 
		FROM orders 
		JOIN users ON orders.user_id = users.id 
		JOIN stores ON orders.store_id = stores.id 
		WHERE orders.user_id = $1 AND orders.id = $2
	`, userID, orderID)

	var order domain.OrderDTO
	if err := rowsOrder.Scan(&order.Id, &order.User, &order.Store, &order.CreatedAt, &order.UpdateAt); err != nil {
		return result, fmt.Errorf("Error scan rows order: %w", err)
	}

	rowsItems, err := tx.Query(ctx, `
		SELECT order_items.id, products.name, order_items.quantity 
		FROM order_items 
		JOIN orders ON order_items.order_id = orders.id 
		JOIN products ON order_items.product_id = products.id
		WHERE order_items.order_id = $1 
	`, orderID)
	if err != nil {
		return result, fmt.Errorf("Error get order items: %w", err)
	}
	defer rowsItems.Close()

	var items []domain.OrderItemsDTO
	for rowsItems.Next() {
		var item domain.OrderItemsDTO

		if err = rowsItems.Scan(&item.Id, &item.Title, &item.Quantity); err != nil {
			return result, fmt.Errorf("Error scan rows order items: %w", err)
		}

		items = append(items, item)
	}

	order.ItemsCount = len(items)

	result.Order = order
	result.Items = items

	return result, nil
}
