package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
)

type ServiceOrderItem struct {
	storage domain.Storage
	order   domain.OrderRepository
	item    domain.OrderItemRepository
}

func NewServiceOrderItem(st domain.Storage, order domain.OrderRepository, item domain.OrderItemRepository) *ServiceOrderItem {
	return &ServiceOrderItem{
		storage: st,
		order:   order,
		item:    item,
	}
}

func (s *ServiceOrderItem) WithTx(ctx context.Context, fn func(q domain.Querier) error) (err error) {
	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = fmt.Errorf("tx err: %v, rollback err: %w", err, rollbackErr)
			}
			return
		}

		if commitErr := tx.Commit(ctx); commitErr != nil {
			err = fmt.Errorf("commit err: %w", commitErr)
		}
	}()

	err = fn(tx)
	return err
}

func (s *ServiceOrderItem) GetAccessibleOrder(ctx context.Context, actor policy.Actor, orderID int) (domain.OrderWithItemDetails, error) {
	logger := logging.LoggerFromContext(ctx)

	order, err := s.order.GetOrder(ctx, s.storage, actor.UserID, orderID)
	if err != nil {
		logger.WarnContext(ctx, "failed to get order for access check", "order_id", orderID, "error", err)
		return domain.OrderWithItemDetails{}, err
	}
	if err := policy.CanAccess(actor, order.Order); err != nil {
		logger.WarnContext(ctx, "access denied to order", "order_id", orderID)
		return domain.OrderWithItemDetails{}, err
	}
	return order, nil
}

func (s *ServiceOrderItem) CreateOrder(ctx context.Context, actor policy.Actor, storeID int) (domain.OrderWithItemDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("store_id", storeID)
	logger.InfoContext(ctx, "creating new order")

	var order domain.OrderWithItemDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		order, err = s.order.CreateOrder(ctx, q, actor.UserID, storeID)
		return err
	}); err != nil {
		logger.ErrorContext(ctx, "failed to create order", "error", err)
		return domain.OrderWithItemDetails{}, err
	}

	logger.InfoContext(ctx, "order created", "order_id", order.Order.ID)
	return order, nil
}

func (s *ServiceOrderItem) GetOrder(ctx context.Context, actor policy.Actor, orderID int) (domain.OrderWithItemDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID)
	logger.InfoContext(ctx, "getting order")

	order, err := s.GetAccessibleOrder(ctx, actor, orderID)
	if err != nil {
		logger.WarnContext(ctx, "failed to get order", "error", err)
		return domain.OrderWithItemDetails{}, err
	}

	logger.InfoContext(ctx, "order retrieved")
	return order, nil
}

func (s *ServiceOrderItem) DeleteOrder(ctx context.Context, actor policy.Actor, orderID int) error {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID)
	logger.InfoContext(ctx, "deleting order")

	_, err := s.GetAccessibleOrder(ctx, actor, orderID)
	if err != nil {
		logger.WarnContext(ctx, "access denied or order not found", "error", err)
		return err
	}

	err = s.WithTx(ctx, func(q domain.Querier) error {
		return s.order.DeleteOrder(ctx, q, actor.UserID, orderID)
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to delete order", "error", err)
		return err
	}

	logger.InfoContext(ctx, "order deleted")
	return nil
}

func (s *ServiceOrderItem) ListOrders(ctx context.Context, actor policy.Actor) ([]domain.OrderDetails, error) {
	logger := logging.LoggerFromContext(ctx)
	logger.InfoContext(ctx, "listing orders")

	orders, err := s.order.ListOrders(ctx, s.storage, actor.UserID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list orders", "error", err)
		return nil, err
	}

	logger.InfoContext(ctx, "orders listed", "count", len(orders))
	return orders, nil
}

func (s *ServiceOrderItem) AddItem(ctx context.Context, actor policy.Actor, orderID, productID, quantity int) (domain.OrderItemDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "product_id", productID, "quantity", quantity)
	logger.InfoContext(ctx, "adding item to order")

	_, err := s.GetAccessibleOrder(ctx, actor, orderID)
	if err != nil {
		logger.WarnContext(ctx, "access denied or order not found", "error", err)
		return domain.OrderItemDetails{}, err
	}
	var item domain.OrderItemDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		if err := s.item.UpsertItem(ctx, q, orderID, item.ProductID, item.Quantity); err != nil {
			return fmt.Errorf("upsert item %d: %w", item.ProductID, err)
		}
		return err
	}); err != nil {
		logger.ErrorContext(ctx, "failed to add item", "error", err)
		return domain.OrderItemDetails{}, err
	}

	logger.InfoContext(ctx, "item added/updated", "new_quantity", item.Quantity)
	return item, nil
}

func (s *ServiceOrderItem) AddListItems(ctx context.Context, actor policy.Actor, orderID int, items []domain.OrderItem) error {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "count", len(items))
	logger.InfoContext(ctx, "adding list items to order")

	_, err := s.GetAccessibleOrder(ctx, actor, orderID)
	if err != nil {
		logger.WarnContext(ctx, "access denied or order not found", "error", err)
		return err
	}

	return s.WithTx(ctx, func(q domain.Querier) error {
		for _, item := range items {
			if err := s.item.UpsertItem(ctx, q, orderID, item.ProductID, item.Quantity); err != nil {
				return fmt.Errorf("upsert item %d: %w", item.ProductID, err)
			}
		}
		logger.InfoContext(ctx, "all items upserted successfully")
		return nil
	})
}

func (s *ServiceOrderItem) UpdateItem(ctx context.Context, actor policy.Actor, orderID, productID, quantity int) (domain.OrderItemDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "product_id", productID, "quantity", quantity)
	logger.InfoContext(ctx, "updating item quantity")

	_, err := s.GetAccessibleOrder(ctx, actor, orderID)
	if err != nil {
		logger.WarnContext(ctx, "access denied or order not found", "error", err)
		return domain.OrderItemDetails{}, err
	}
	var item domain.OrderItemDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		item, err = s.item.UpdateItem(ctx, q, orderID, productID, quantity)
		return err
	}); err != nil {
		logger.ErrorContext(ctx, "failed to update item", "error", err)
		return domain.OrderItemDetails{}, err
	}

	logger.InfoContext(ctx, "item updated", "new_quantity", item.Quantity)
	return item, nil
}

func (s *ServiceOrderItem) DeleteItem(ctx context.Context, actor policy.Actor, orderID, productID int) error {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "product_id", productID)
	logger.InfoContext(ctx, "deleting item from order")

	_, err := s.GetAccessibleOrder(ctx, actor, orderID)
	if err != nil {
		logger.WarnContext(ctx, "access denied or order not found", "error", err)
		return err
	}
	err = s.WithTx(ctx, func(q domain.Querier) error {
		return s.item.DeleteItem(ctx, q, orderID, productID)
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to delete item", "error", err)
		return err
	}

	logger.InfoContext(ctx, "item deleted")
	return nil
}
