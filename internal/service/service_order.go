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

// WithTx выполняет функцию в транзакции.
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

// AccessReadOrder проверяет доступ к чтению заказа и возвращает его с позициями.
func (s *ServiceOrderItem) AccessReadOrder(ctx context.Context, actor policy.Actor, orderID int) (domain.OrderWithItemDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID)
	logger.InfoContext(ctx, "checking read access for order")

	// 1. Получение заказа из БД
	order, err := s.order.GetOrderByID(ctx, s.storage, orderID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get order", "error", err)
		return domain.OrderWithItemDetails{}, fmt.Errorf("get order: %w", err)
	}

	// 2. Проверка прав на чтение через policy
	if err := policy.CanGroupAccessForReading(actor, order); err != nil {
		logger.WarnContext(ctx, "read access denied", "error", err)
		return domain.OrderWithItemDetails{}, err
	}

	logger.InfoContext(ctx, "read access granted")
	return order, nil
}

// AccessWriteOrder проверяет доступ к изменению заказа.
func (s *ServiceOrderItem) AccessWriteOrder(ctx context.Context, actor policy.Actor, orderID int) error {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID)
	logger.InfoContext(ctx, "checking write access for order")

	// 1. Получение заказа из БД
	order, err := s.order.GetOrderByID(ctx, s.storage, orderID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get order", "error", err)
		return fmt.Errorf("get order: %w", err)
	}

	// 2. Проверка прав на изменение через policy
	if err := policy.CanGroupAccessForModify(actor, order); err != nil {
		logger.WarnContext(ctx, "write access denied", "error", err)
		return err
	}

	logger.InfoContext(ctx, "write access granted")
	return nil
}

// CreateOrder создаёт новый заказ в группе актора.
func (s *ServiceOrderItem) CreateOrder(ctx context.Context, actor policy.Actor, storeID int) (domain.OrderCreateDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("store_id", storeID, "group_id", actor.GroupID)
	logger.InfoContext(ctx, "creating new order")

	var order domain.OrderCreateDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		// Создание заказа с userID и groupID актора
		order, err = s.order.CreateOrder(ctx, q, actor.UserID, storeID, actor.GroupID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to create order", "error", err)
			return fmt.Errorf("create order: %w", err)
		}
		return nil
	}); err != nil {
		return domain.OrderCreateDetails{}, fmt.Errorf("create order: %w", err)
	}

	logger.InfoContext(ctx, "order created successfully", "order_id", order.ID)
	return order, nil
}

// GetOrder возвращает заказ по ID с проверкой прав на чтение.
func (s *ServiceOrderItem) GetOrder(ctx context.Context, actor policy.Actor, orderID int) (domain.OrderWithItemDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID)
	logger.InfoContext(ctx, "getting order")

	order, err := s.AccessReadOrder(ctx, actor, orderID)
	if err != nil {
		return domain.OrderWithItemDetails{}, fmt.Errorf("access read order: %w", err)
	}

	logger.InfoContext(ctx, "order retrieved successfully")
	return order, nil
}

// DeleteOrder удаляет заказ с проверкой прав на изменение.
func (s *ServiceOrderItem) DeleteOrder(ctx context.Context, actor policy.Actor, orderID int) error {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID)
	logger.InfoContext(ctx, "deleting order")

	// 1. Проверка прав на запись
	if err := s.AccessWriteOrder(ctx, actor, orderID); err != nil {
		return fmt.Errorf("access write order: %w", err)
	}

	// 2. Удаление заказа в транзакции
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		if err := s.order.DeleteOrderByID(ctx, q, orderID); err != nil {
			logger.ErrorContext(ctx, "failed to delete order", "error", err)
			return fmt.Errorf("delete order: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("delete order: %w", err)
	}

	logger.InfoContext(ctx, "order deleted successfully")
	return nil
}

// ListOrders возвращает список заказов пользователя в его группе.
func (s *ServiceOrderItem) ListOrders(ctx context.Context, actor policy.Actor) ([]domain.OrderDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("user_id", actor.UserID, "group_id", actor.GroupID)
	logger.InfoContext(ctx, "listing orders")

	orders, err := s.order.ListOrders(ctx, s.storage, actor.UserID, actor.GroupID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list orders", "error", err)
		return []domain.OrderDetails{}, fmt.Errorf("list orders: %w", err)
	}

	logger.InfoContext(ctx, "orders listed successfully", "count", len(orders))
	return orders, nil
}

// ---------------------------------------------------------------------------------
// --------------------------------------ITEMS--------------------------------------
// ---------------------------------------------------------------------------------

// AddItem добавляет позицию в заказ (или обновляет количество). Проверяет права на запись заказа.
func (s *ServiceOrderItem) AddItem(ctx context.Context, actor policy.Actor, orderID, productID, unitID, quantity int) (domain.OrderItemDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "product_id", productID, "unit_id", unitID, "quantity", quantity)
	logger.InfoContext(ctx, "adding item to order")

	// 1. Проверка прав на запись заказа
	if err := s.AccessWriteOrder(ctx, actor, orderID); err != nil {
		return domain.OrderItemDetails{}, fmt.Errorf("access write order: %w", err)
	}

	var item domain.OrderItemDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		// 2. Добавление или обновление позиции
		item, err = s.item.AddItem(ctx, q, orderID, productID, unitID, quantity, actor.GroupID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to add item", "error", err)
			return fmt.Errorf("add item: %w", err)
		}
		return nil
	}); err != nil {
		return domain.OrderItemDetails{}, fmt.Errorf("add item: %w", err)
	}

	logger.InfoContext(ctx, "item added/updated successfully", "new_quantity", item.Quantity)
	return item, nil
}

// AddListItems добавляет несколько позиций в заказ (upsert). Проверяет права на запись заказа.
func (s *ServiceOrderItem) AddListItems(ctx context.Context, actor policy.Actor, orderID int, items []domain.OrderItemDetails) error {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "count", len(items))
	logger.InfoContext(ctx, "adding list items to order")

	// 1. Проверка прав на запись заказа
	if err := s.AccessWriteOrder(ctx, actor, orderID); err != nil {
		return fmt.Errorf("access write order: %w", err)
	}

	// 2. Upsert каждой позиции в транзакции
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		for _, item := range items {
			_, err := s.item.GetItemByOrderAndProduct(ctx, q, orderID, item.ProductID)
			if err != nil {
				if domain.IsNotFound(err) {
					// Если позиции нет — добавляем
					if _, err := s.item.AddItem(ctx, q, orderID, item.ProductID, item.UnitID, item.Quantity, actor.GroupID); err != nil {
						return fmt.Errorf("add item %d: %w", item.ProductID, err)
					}
					continue
				}
				return fmt.Errorf("get item %d: %w", item.ProductID, err)
			}
			// Если есть — обновляем
			if _, err := s.item.UpdateItem(ctx, q, orderID, item.ProductID, domain.OrderItemUpdate{UnitID: &item.UnitID, Quantity: &item.Quantity}); err != nil {
				return fmt.Errorf("update item %d: %w", item.ProductID, err)
			}
		}
		return nil
	}); err != nil {
		logger.ErrorContext(ctx, "failed to add list items", "error", err)
		return fmt.Errorf("add list items: %w", err)
	}

	logger.InfoContext(ctx, "all items upserted successfully")
	return nil
}

// UpdateListItems обновляет список позиций (аналогично AddListItems — upsert). Проверяет права на запись заказа.
func (s *ServiceOrderItem) UpdateListItems(ctx context.Context, actor policy.Actor, orderID int, items []domain.OrderItemDetails) error {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "count", len(items))
	logger.InfoContext(ctx, "updating list items")

	// 1. Проверка прав на запись заказа
	if err := s.AccessWriteOrder(ctx, actor, orderID); err != nil {
		return fmt.Errorf("access write order: %w", err)
	}

	// 2. Upsert каждой позиции в транзакции
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		for _, item := range items {
			_, err := s.item.GetItemByOrderAndProduct(ctx, q, orderID, item.ProductID)
			if err != nil {
				if domain.IsNotFound(err) {
					if _, err := s.item.AddItem(ctx, q, orderID, item.ProductID, item.UnitID, item.Quantity, actor.GroupID); err != nil {
						return fmt.Errorf("add item %d: %w", item.ProductID, err)
					}
					continue
				}
				return fmt.Errorf("get item %d: %w", item.ProductID, err)
			}
			if _, err := s.item.UpdateItem(ctx, q, orderID, item.ProductID, domain.OrderItemUpdate{UnitID: &item.UnitID, Quantity: &item.Quantity}); err != nil {
				return fmt.Errorf("update item %d: %w", item.ProductID, err)
			}
		}
		return nil
	}); err != nil {
		logger.ErrorContext(ctx, "failed to update list items", "error", err)
		return fmt.Errorf("update list items: %w", err)
	}

	logger.InfoContext(ctx, "all items upserted successfully")
	return nil
}

// UpdateItem обновляет количество или единицу измерения позиции. Проверяет права на запись заказа.
func (s *ServiceOrderItem) UpdateItem(ctx context.Context, actor policy.Actor, orderID, productID int, updateOrder domain.OrderItemUpdate) (domain.OrderItemDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "product_id", productID)
	logger.InfoContext(ctx, "updating item")

	// 1. Проверка прав на запись заказа
	if err := s.AccessWriteOrder(ctx, actor, orderID); err != nil {
		return domain.OrderItemDetails{}, fmt.Errorf("access write order: %w", err)
	}

	var item domain.OrderItemDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		// 2. Обновление позиции в БД
		item, err = s.item.UpdateItem(ctx, q, orderID, productID, updateOrder)
		if err != nil {
			logger.ErrorContext(ctx, "failed to update item", "error", err)
			return fmt.Errorf("update item: %w", err)
		}
		return nil
	}); err != nil {
		return domain.OrderItemDetails{}, fmt.Errorf("update item: %w", err)
	}

	logger.InfoContext(ctx, "item updated successfully", "new_quantity", item.Quantity)
	return item, nil
}

// DeleteItem удаляет позицию из заказа. Проверяет права на запись заказа.
func (s *ServiceOrderItem) DeleteItem(ctx context.Context, actor policy.Actor, orderID, productID int) error {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "product_id", productID)
	logger.InfoContext(ctx, "deleting item from order")

	// 1. Проверка прав на запись заказа
	if err := s.AccessWriteOrder(ctx, actor, orderID); err != nil {
		return fmt.Errorf("access write order: %w", err)
	}

	// 2. Удаление позиции в транзакции
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		if err := s.item.DeleteItemByOrderAndProduct(ctx, q, orderID, productID); err != nil {
			logger.ErrorContext(ctx, "failed to delete item", "error", err)
			return fmt.Errorf("delete item: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	logger.InfoContext(ctx, "item deleted successfully")
	return nil
}
