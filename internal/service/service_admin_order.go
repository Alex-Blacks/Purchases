package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
)

type ServiceAdminOrderItem struct {
	storage domain.Storage
	order   domain.OrderRepository
	item    domain.OrderItemRepository
}

func NewServiceAdminOrderItem(st domain.Storage, order domain.OrderRepository, item domain.OrderItemRepository) *ServiceAdminOrderItem {
	return &ServiceAdminOrderItem{
		storage: st,
		order:   order,
		item:    item,
	}
}

func (s *ServiceAdminOrderItem) WithTx(ctx context.Context, fn func(q domain.Querier) error) (err error) {
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

func (s *ServiceAdminOrderItem) CreateOrder(ctx context.Context, UserID, storeID, groupID int) (domain.OrderCreateDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("store_id", storeID)
	logger.InfoContext(ctx, "creating new order")

	var order domain.OrderCreateDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		order, err = s.order.CreateOrder(ctx, q, UserID, storeID, groupID)
		return err
	}); err != nil {
		logger.ErrorContext(ctx, "failed to create order", "error", err)
		return domain.OrderCreateDetails{}, err
	}

	logger.InfoContext(ctx, "order created", "order_id", order.ID)
	return order, nil
}

func (s *ServiceAdminOrderItem) GetOrder(ctx context.Context, orderID int) (domain.OrderWithItemDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID)
	logger.InfoContext(ctx, "getting order")

	order, err := s.order.GetOrderByID(ctx, s.storage, orderID)
	if err != nil {
		return domain.OrderWithItemDetails{}, err
	}
	logger.InfoContext(ctx, "order retrieved")
	return order, nil
}

func (s *ServiceAdminOrderItem) DeleteOrder(ctx context.Context, orderID int) error {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID)
	logger.InfoContext(ctx, "deleting order")

	if err := s.WithTx(ctx, func(q domain.Querier) error {
		return s.order.DeleteOrderByID(ctx, q, orderID)
	}); err != nil {
		logger.ErrorContext(ctx, "failed to delete order", "error", err)
		return err
	}

	logger.InfoContext(ctx, "order deleted")
	return nil
}

func (s *ServiceAdminOrderItem) ListOrders(ctx context.Context) ([]domain.OrderDetails, error) {
	logger := logging.LoggerFromContext(ctx)
	logger.InfoContext(ctx, "listing orders")

	orders, err := s.order.ListAdminOrders(ctx, s.storage)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list orders", "error", err)
		return nil, err
	}

	logger.InfoContext(ctx, "orders listed", "count", len(orders))
	return orders, nil
}

// ---------------------------------------------------------------------------------
// --------------------------------------ITEMS--------------------------------------
// ---------------------------------------------------------------------------------

func (s *ServiceAdminOrderItem) AddItem(ctx context.Context, orderID, productID, unitID, quantity, groupID int) (domain.OrderItemDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "product_id", productID, "unitID", unitID, "quantity", quantity)
	logger.InfoContext(ctx, "adding item to order")

	var item domain.OrderItemDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		if item, err = s.item.AddItem(ctx, q, orderID, productID, unitID, quantity, groupID); err != nil {
			return fmt.Errorf("add item: %w", err)
		}
		return err
	}); err != nil {
		logger.ErrorContext(ctx, "failed to add item", "error", err)
		return domain.OrderItemDetails{}, err
	}

	logger.InfoContext(ctx, "item added/updated", "new_quantity", item.Quantity)
	return item, nil
}

func (s *ServiceAdminOrderItem) AddListItems(ctx context.Context, orderID int, items []domain.OrderItemDetails, groupID int) error {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "count", len(items))
	logger.InfoContext(ctx, "adding list items to order")

	return s.WithTx(ctx, func(q domain.Querier) error {
		for _, item := range items {
			if _, err := s.item.GetItemByOrderAndProduct(ctx, q, orderID, item.ProductID); err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					if _, err := s.item.AddItem(ctx, q, orderID, item.ProductID, item.UnitID, item.Quantity, groupID); err != nil {
						return fmt.Errorf("add item %d: %w", item.ProductID, err)
					}
					continue
				}
				return fmt.Errorf("get item %d: %w", item.ProductID, err)
			}
			if _, err := s.item.UpdateItem(ctx, q, orderID, item.ProductID, domain.OrderItemUpdate{UnitID: &item.UnitID, Quantity: &item.Quantity}); err != nil {
				return fmt.Errorf("upsert item %d: %w", item.ProductID, err)
			}
		}
		logger.InfoContext(ctx, "all items upserted successfully")
		return nil
	})
}

func (s *ServiceAdminOrderItem) UpdateListItems(ctx context.Context, orderID int, items []domain.OrderItemDetails, groupID int) error {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "count", len(items))
	logger.InfoContext(ctx, "updating list items")

	return s.WithTx(ctx, func(q domain.Querier) error {
		for _, item := range items {
			if _, err := s.item.GetItemByOrderAndProduct(ctx, q, orderID, item.ProductID); err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					if _, err := s.item.AddItem(ctx, q, orderID, item.ProductID, item.UnitID, item.Quantity, groupID); err != nil {
						return fmt.Errorf("add item %d: %w", item.ProductID, err)
					}
					continue
				}
				return fmt.Errorf("get item %d: %w", item.ProductID, err)
			}
			if _, err := s.item.UpdateItem(ctx, q, orderID, item.ProductID, domain.OrderItemUpdate{UnitID: &item.UnitID, Quantity: &item.Quantity}); err != nil {
				return fmt.Errorf("upsert item %d: %w", item.ProductID, err)
			}
		}
		logger.InfoContext(ctx, "all items upserted successfully")
		return nil
	})
}

func (s *ServiceAdminOrderItem) UpdateItem(ctx context.Context, orderID, productID int, updateOrder domain.OrderItemUpdate) (domain.OrderItemDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "product_id", productID, "updateOrder", updateOrder)
	logger.InfoContext(ctx, "updating item quantity")

	var item domain.OrderItemDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		item, err = s.item.UpdateItem(ctx, q, orderID, productID, updateOrder)
		return err
	}); err != nil {
		logger.ErrorContext(ctx, "failed to update item", "error", err)
		return domain.OrderItemDetails{}, err
	}

	logger.InfoContext(ctx, "item updated", "new_quantity", item.Quantity)
	return item, nil
}

func (s *ServiceAdminOrderItem) DeleteItem(ctx context.Context, orderID, productID int) error {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "product_id", productID)
	logger.InfoContext(ctx, "deleting item from order")

	err := s.WithTx(ctx, func(q domain.Querier) error {
		return s.item.DeleteItemByOrderAndProduct(ctx, q, orderID, productID)
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to delete item", "error", err)
		return err
	}

	logger.InfoContext(ctx, "item deleted")
	return nil
}
