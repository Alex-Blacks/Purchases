package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type OrderRepo struct{}

func NewOrderRepo() *OrderRepo {
	return &OrderRepo{}
}

type OrderItemRepo struct {
}

func NewOrderItemRepo() *OrderItemRepo {
	return &OrderItemRepo{}
}

func (r *OrderRepo) CreateOrder(ctx context.Context, q domain.Querier, userID, storeID int) (domain.OrderWithItemDetails, error) {
	var order domain.OrderWithItemDetails
	if err := q.QueryRow(ctx, `INSERT INTO orders(user_id, store_id) VALUES ($1, $2) RETURNING id`, userID, storeID).Scan(&order.Order.ID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.OrderWithItemDetails{}, domain.ErrAlreadyExists
		}
		return domain.OrderWithItemDetails{}, fmt.Errorf("create order failed: %w", err)
	}

	orderOutput, err := r.GetOrder(ctx, q, userID, order.Order.ID)
	if err != nil {
		return domain.OrderWithItemDetails{}, fmt.Errorf("get order: %w", err)
	}
	return orderOutput, nil
}

func (r *OrderRepo) GetOrder(ctx context.Context, q domain.Querier, userID, orderID int) (domain.OrderWithItemDetails, error) {
	var result domain.OrderWithItemDetails
	rowsOrder := q.QueryRow(ctx, `
		SELECT o.id, o.user_id, s.name, o.created_at, o.updated_at 
		FROM orders o
		JOIN stores s ON o.store_id = s.id 
		WHERE o.user_id = $1 AND o.id = $2
	`, userID, orderID)

	var order domain.OrderDetails
	if err := rowsOrder.Scan(&order.ID, &order.UserID, &order.Store, &order.CreatedAt, &order.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return result, domain.ErrNotFound
		}
		return result, fmt.Errorf("scan rows order: %w", err)
	}

	rowsItems, err := q.Query(ctx, `
		SELECT oi.id, oi.product_id, p.title, oi.quantity 
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = $1 
	`, orderID)
	if err != nil {
		return result, fmt.Errorf("get order items: %w", err)
	}
	defer rowsItems.Close()

	var items []domain.OrderItemDetails
	for rowsItems.Next() {
		var item domain.OrderItemDetails

		if err = rowsItems.Scan(&item.ID, &item.ProductID, &item.Title, &item.Quantity); err != nil {
			return result, fmt.Errorf("scan rows order items: %w", err)
		}

		items = append(items, item)
	}
	if err := rowsItems.Err(); err != nil {
		return result, fmt.Errorf("iteration failed: %w", err)
	}

	order.ItemsCount = len(items)

	result.Order = order
	result.Items = items

	return result, nil
}

func (r *OrderRepo) DeleteOrder(ctx context.Context, q domain.Querier, userID, orderID int) error {
	var OrderID int
	if err := q.QueryRow(ctx, `DELETE FROM orders WHERE orders.id = $1 AND orders.user_id = $2 RETURNING id`, orderID, userID).Scan(&OrderID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("delete order: %w", err)
	}

	return nil
}

func (r *OrderRepo) ListOrders(ctx context.Context, q domain.Querier, userID int) ([]domain.OrderDetails, error) {
	rows, err := q.Query(ctx, `
		SELECT 
			o.id, u.name, s.name, o.created_at, o.updated_at, 
			COUNT(oi.id) AS items_quantity
		FROM orders o
		JOIN users u ON o.user_id = u.id
		JOIN stores s ON o.store_id = s.id
		LEFT JOIN order_items oi ON oi.order_id = o.id
		WHERE o.user_id = $1
		GROUP BY o.id, u.name, s.name, o.created_at, o.updated_at
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("query order: %w", err)
	}
	defer rows.Close()

	var lists []domain.OrderDetails
	for rows.Next() {
		var list domain.OrderDetails

		if err := rows.Scan(&list.ID, &list.UserID, &list.Store, &list.CreatedAt, &list.UpdatedAt, &list.ItemsCount); err != nil {
			return nil, fmt.Errorf("scan orders: %w", err)
		}

		lists = append(lists, list)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration failed: %w", err)
	}

	return lists, nil
}

func (r *OrderItemRepo) AddItem(ctx context.Context, q domain.Querier, orderID, productID, quantity int) (domain.OrderItemDetails, error) {
	var item domain.OrderItemDetails
	if err := q.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO order_items(order_id, product_id, quantity) 
			VALUES ($1,$2,$3) 
			RETURNING id, product_id, quantity
		)
		SELECT i.id, i.product_id, p.title, i.quantity
		FROM inserted i
		JOIN products p ON i.product_id = p.id 
	`, orderID, productID, quantity).Scan(&item.ID, &item.ProductID, &item.Title, &item.Quantity); err != nil {
		var pgErr *pgconn.PgError
		if errors.Is(err, pgx.ErrNoRows) {
			return item, domain.ErrNotFound
		}
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return item, domain.ErrAlreadyExists
		}
		return item, fmt.Errorf("add item: %w", err)
	}

	return item, nil
}

func (r *OrderItemRepo) UpdateItem(ctx context.Context, q domain.Querier, orderID, productID, quantity int) (domain.OrderItemDetails, error) {
	var item domain.OrderItemDetails
	if err := q.QueryRow(ctx, `
		WITH updated AS (
			UPDATE order_items
			SET quantity = $1
			WHERE order_id = $2 AND product_id = $3
			RETURNING id, product_id, quantity
		)
		SELECT u.id, u.product_id, p.title, u.quantity 
		FROM updated u	
		JOIN products p ON u.product_id = p.id 
	`, quantity, orderID, productID).Scan(&item.ID, &item.ProductID, &item.Title, &item.Quantity); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return item, domain.ErrNotFound
		}
		return item, fmt.Errorf("update item: %w", err)
	}

	return item, nil
}

func (r *OrderItemRepo) DeleteItem(ctx context.Context, q domain.Querier, orderID, productID int) error {
	var item int
	if err := q.QueryRow(ctx, `DELETE FROM order_items WHERE order_items.order_id = $1 AND order_items.product_id = $2 RETURNING id`, orderID, productID).Scan(&item); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("deleted item: %w", err)
	}

	return nil
}
