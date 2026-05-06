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
	st *Storage
}

func NewOrderRepo(st *Storage) *OrderRepo {
	if st == nil {
		panic("storage is nil")
	}
	return &OrderRepo{st: st}
}

type OrderItemRepo struct {
	st *Storage
}

func NewOrderItemRepo(st *Storage) *OrderItemRepo {
	if st == nil {
		panic("storage is nil")
	}
	return &OrderItemRepo{st: st}
}

func (r *OrderRepo) CreateOrder(ctx context.Context, q domain.Querier, userID, storeID int) (int, error) {
	var id int
	if err := q.QueryRow(ctx, `INSERT INTO orders(user_id, store_id) VALUES ($1, $2) RETURNING id`, userID, storeID).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, domain.ErrAlreadyExists
		}
		return 0, fmt.Errorf("create order failed: %w", err)
	}

	return id, nil
}

func (r *OrderRepo) GetOrder(ctx context.Context, q domain.Querier, userID, orderID int) (domain.OrderWithItemsDTO, error) {
	var result domain.OrderWithItemsDTO
	rowsOrder := q.QueryRow(ctx, `
		SELECT orders.id, users.name, stores.name, orders.created_at, orders.updated_at 
		FROM orders 
		JOIN users ON orders.user_id = users.id 
		JOIN stores ON orders.store_id = stores.id 
		WHERE orders.user_id = $1 AND orders.id = $2
	`, userID, orderID)

	var order domain.OrderDTO
	if err := rowsOrder.Scan(&order.Id, &order.User, &order.Store, &order.CreatedAt, &order.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return result, domain.ErrNotFound
		}
		return result, fmt.Errorf("Error scan rows order: %w", err)
	}

	rowsItems, err := q.Query(ctx, `
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
	if err := rowsItems.Err(); err != nil {
		return result, fmt.Errorf("Error iteration failed: %w", err)
	}

	order.ItemsCount = len(items)

	result.Order = order
	result.Items = items

	return result, nil
}

func (r *OrderRepo) DeleteOrder(ctx context.Context, q domain.Querier, userID, orderID int) error {
	tag, err := q.Exec(ctx, `DELETE FROM orders WHERE orders.id = $1 AND orders.user_id = $2`, orderID, userID)
	if err != nil {
		return fmt.Errorf("Error delete order: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *OrderRepo) ListOrders(ctx context.Context, q domain.Querier, userID int) ([]domain.OrderDTO, error) {
	rows, err := q.Query(ctx, `
		SELECT 
			o.id, u.name, s.name, o.created_at, o.updated_at, 
			COUNT(oi.id) AS items_count
		FROM orders o
		JOIN users u ON o.user_id = u.id
		JOIN stores s ON o.store_id = s.id
		LEFT JOIN order_items oi ON oi.order_id = o.id
		WHERE o.user_id = $1
		GROUP BY o.id, u.name, s.name, o.created_at, o.updated_at
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("Error query order: %w", err)
	}

	var lists []domain.OrderDTO
	for rows.Next() {
		var list domain.OrderDTO

		if err := rows.Scan(&list.Id, &list.User, &list.Store, &list.CreatedAt, &list.UpdatedAt, &list.ItemsCount); err != nil {
			return nil, fmt.Errorf("Error scan orders: %w", err)
		}

		lists = append(lists, list)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error iteration failed: %w", err)
	}

	return lists, nil
}

func (r *OrderItemRepo) AddItem(ctx context.Context, q domain.Querier, orderID, productID, qty int) error {
	if _, err := q.Exec(ctx, `INSERT INTO order_items(order_id, product_id, quantity) VALUES ($1,$2,$3)`, orderID, productID, qty); err != nil {
		return fmt.Errorf("Error add item: %w", err)
	}

	return nil
}

func (r *OrderItemRepo) UpdateItem(ctx context.Context, q domain.Querier, orderID, productID, qty int) error {
	tag, err := q.Exec(ctx, `
		UPDATE order_items o
		SET o.quantity = $1
		WHERE o.order_id = $2 AND o.product_id = $3
	`, qty, orderID, productID)
	if err != nil {
		return fmt.Errorf("Error update item: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *OrderItemRepo) DeleteItem(ctx context.Context, q domain.Querier, orderID, productID int) error {
	tag, err := q.Exec(ctx, `DELETE FROM order_items WHERE order_items.order_id = $1 AND order_items.product_id = $2`, orderID, productID)
	if err != nil {
		return fmt.Errorf("Error deleted item: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *OrderItemRepo) ClearOrder(ctx context.Context, q domain.Querier, orderID int) error {
	_, err := q.Exec(ctx, `DELETE FROM order_items WHERE order_items.order_id = $1`, orderID)
	if err != nil {
		return fmt.Errorf("Error delete items order: %w", err)
	}

	return nil
}
