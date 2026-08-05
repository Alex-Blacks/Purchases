package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

func (r *OrderRepo) CreateOrder(ctx context.Context, q domain.Querier, userID, storeID, groupID int) (domain.OrderCreateDetails, error) {
	var order domain.OrderCreateDetails
	if err := q.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO orders(user_id, store_id, group_id) 
			VALUES ($1, $2, $3) 
			RETURNING id, user_id, store_id, group_id, created_at, updated_at
		)
		SELECT i.id, i.user_id, u.name, i.store_id, s.name, i.group_id, g.name, i.created_at, i.updated_at 
		FROM inserted i
		JOIN users u ON i.user_id = u.id
		JOIN stores s ON i.store_id = s.id
		JOIN groups g ON i.group_id = g.id
	`, userID, storeID, groupID).Scan(&order.ID, &order.UserID, &order.User, &order.StoreID, &order.Store, &order.GroupID, &order.Group, &order.CreatedAt, &order.UpdatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolation:
				return domain.OrderCreateDetails{}, domain.ErrAlreadyExists
			case pgForeignKeyViolation:
				return domain.OrderCreateDetails{}, domain.ErrConflict
			}
		}
		return domain.OrderCreateDetails{}, fmt.Errorf("create order: %w", err)
	}

	return order, nil
}

func (r *OrderRepo) GetOrderByID(ctx context.Context, q domain.Querier, orderID int) (domain.OrderWithItemDetails, error) {
	var result domain.OrderWithItemDetails
	rowsOrder := q.QueryRow(ctx, `
		SELECT o.id, o.user_id, u.name, o.store_id, s.name, o.group_id, g.name, o.created_at, o.updated_at 
		FROM orders o
		JOIN stores s ON o.store_id = s.id 
		JOIN users u ON o.user_id = u.id
		JOIN groups g ON o.group_id = g.id
		WHERE o.id = $1
	`, orderID)

	var order domain.OrderDetails
	if err := rowsOrder.Scan(&order.ID, &order.UserID, &order.User, &order.StoreID, &order.Store, &order.GroupID, &order.Group, &order.CreatedAt, &order.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return result, domain.ErrNotFound
		}
		return domain.OrderWithItemDetails{}, fmt.Errorf("scan rows order: %w", err)
	}

	rowsItems, err := q.Query(ctx, `
		SELECT oi.id, oi.order_id, oi.product_id, p.title, oi.unit_id, u.short_name, oi.quantity, oi.group_id, g.name 
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		JOIN units u ON oi.unit_id = u.id
		JOIN groups g ON oi.group_id = g.id
		WHERE oi.order_id = $1 
	`, orderID)
	if err != nil {
		return domain.OrderWithItemDetails{}, fmt.Errorf("get order items: %w", err)
	}
	defer rowsItems.Close()

	var items []domain.OrderItemDetails
	for rowsItems.Next() {
		var item domain.OrderItemDetails
		if err = rowsItems.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Title, &item.UnitID, &item.Unit, &item.Quantity, &item.GroupID, &item.Group); err != nil {
			return domain.OrderWithItemDetails{}, fmt.Errorf("scan rows order items: %w", err)
		}

		items = append(items, item)
	}
	if err := rowsItems.Err(); err != nil {
		return domain.OrderWithItemDetails{}, fmt.Errorf("iteration failed: %w", err)
	}

	order.ItemsCount = len(items)

	result.Order = order
	result.Items = items

	return result, nil
}

func (r *OrderRepo) DeleteOrderByID(ctx context.Context, q domain.Querier, orderID int) error {
	var id int
	if err := q.QueryRow(ctx, `DELETE FROM orders WHERE orders.id = $1 RETURNING id`, orderID).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("delete order: %w", err)
	}

	return nil
}

func (r *OrderRepo) ListOrders(ctx context.Context, q domain.Querier, userID int, groupID int) ([]domain.OrderDetails, error) {
	rows, err := q.Query(ctx, `
		SELECT 
			o.id, o.user_id, u.name, o.store_id, s.name, o.group_id, g.name, o.created_at, o.updated_at, 
			COUNT(oi.id) AS items_quantity
		FROM orders o
		JOIN stores s ON o.store_id = s.id
		JOIN users u ON o.user_id = u.id
		JOIN groups g ON o.group_id = g.id
		LEFT JOIN order_items oi ON oi.order_id = o.id
		WHERE o.user_id = $1 OR o.group_id = $2
		GROUP BY o.id, o.user_id, u.name, o.store_id, s.name, o.group_id, g.name, o.created_at, o.updated_at
	`, userID, groupID)
	if err != nil {
		return nil, fmt.Errorf("query order: %w", err)
	}
	defer rows.Close()

	var lists []domain.OrderDetails
	for rows.Next() {
		var list domain.OrderDetails
		if err := rows.Scan(&list.ID, &list.UserID, &list.User, &list.StoreID, &list.Store, &list.GroupID, &list.Group, &list.CreatedAt, &list.UpdatedAt, &list.ItemsCount); err != nil {
			return nil, fmt.Errorf("scan orders: %w", err)
		}

		lists = append(lists, list)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration failed: %w", err)
	}

	return lists, nil
}

func (r *OrderItemRepo) GetItemByOrderAndProduct(ctx context.Context, q domain.Querier, orderID, productID int) (domain.OrderItemDetails, error) {
	var item domain.OrderItemDetails
	if err := q.QueryRow(ctx, `
		SELECT oi.id, oi.order_id, oi.product_id, p.title, oi.unit_id, u.short_name, oi.quantity, oi.group_id, g.name
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		JOIN units u ON oi.unit_id = u.id
		JOIN groups g ON oi.group_id = g.id
		WHERE oi.order_id = $1 AND oi.product_id = $2
	`, orderID, productID).Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Title, &item.UnitID, &item.Unit, &item.Quantity, &item.GroupID, &item.Group); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.OrderItemDetails{}, domain.ErrNotFound
		}
		return domain.OrderItemDetails{}, fmt.Errorf("get item by order and product: %w", err)
	}

	return item, nil
}

func (r *OrderItemRepo) AddItem(ctx context.Context, q domain.Querier, orderID, productID, unitID, quantity int, groupID int) (domain.OrderItemDetails, error) {
	var item domain.OrderItemDetails
	if err := q.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO order_items(order_id, product_id, unit_id, quantity, group_id) 
			VALUES ($1,$2,$3,$4,$5) 
			RETURNING id, order_id, product_id, unit_id, quantity, group_id
		)
		SELECT i.id, i.order_id, i.product_id, p.title, i.unit_id, u.short_name, i.quantity, i.group_id, g.name
		FROM inserted i
		JOIN products p ON i.product_id = p.id 
		JOIN units u ON i.unit_id = u.id 
		JOIN groups g ON i.group_id = g.id
	`, orderID, productID, unitID, quantity, groupID).Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Title, &item.UnitID, &item.Unit, &item.Quantity, &item.GroupID, &item.Group); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.OrderItemDetails{}, domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.OrderItemDetails{}, domain.ErrAlreadyExists
		}
		return domain.OrderItemDetails{}, fmt.Errorf("add item: %w", err)
	}

	return item, nil
}

func (r *OrderItemRepo) DeleteItemByID(ctx context.Context, q domain.Querier, orderID, productID int) error {
	var item int
	if err := q.QueryRow(ctx, `DELETE FROM order_items WHERE order_items.order_id = $1 AND order_items.product_id = $2 RETURNING id`, orderID, productID).Scan(&item); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
			return domain.ErrConflict
		}
		return fmt.Errorf("deleted item: %w", err)
	}

	return nil
}

func (r *OrderItemRepo) DeleteAllItems(ctx context.Context, q domain.Querier, orderID int) error {
	tag, err := q.Exec(ctx, `DELETE FROM order_items WHERE order_items.order_id = $1`, orderID)
	if err != nil {
		return fmt.Errorf("deleted all items: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *OrderItemRepo) UpdateItem(ctx context.Context, q domain.Querier, orderID, productID int, updateItem domain.OrderItemUpdate) (domain.OrderItemDetails, error) {
	var item domain.OrderItemDetails
	args := []any{orderID, productID}
	setParts := []string{}
	argPos := 3

	if updateItem.Quantity != nil && *updateItem.Quantity >= 1 {
		setParts = append(setParts, fmt.Sprintf("quantity = $%d", argPos))
		args = append(args, *updateItem.Quantity)
		argPos++
	}
	if updateItem.UnitID != nil && *updateItem.UnitID >= 1 {
		setParts = append(setParts, fmt.Sprintf("unit_id = $%d", argPos))
		args = append(args, *updateItem.UnitID)
		argPos++
	}

	set := strings.Join(setParts, ", ")
	if strings.TrimSpace(set) == "" {
		return domain.OrderItemDetails{}, domain.ErrNoFieldsToUpdate
	}

	if err := q.QueryRow(ctx, `
		UPDATE order_items oi
		SET `+set+`
		FROM products p
		JOIN units u ON oi.unit_id = u.id
		JOIN groups g ON oi.group_id = g.id
		WHERE oi.order_id = $1 AND oi.product_id = $2
		RETURNING oi.id, oi.order_id, oi.product_id, p.title, oi.unit_id, u.short_name, oi.quantity, oi.group_id, g.name
	`, args...).Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Title, &item.UnitID, &item.Unit, &item.Quantity, &item.GroupID, &item.Group); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.OrderItemDetails{}, domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolation:
				return domain.OrderItemDetails{}, domain.ErrAlreadyExists
			case pgForeignKeyViolation:
				return domain.OrderItemDetails{}, domain.ErrConflict
			}
		}
		return domain.OrderItemDetails{}, fmt.Errorf("update order items: %w", err)
	}
	return item, nil
}
