package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
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

// WithTx выполняет функцию в транзакции.
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

// CreateOrder создаёт новый заказ. Доступно только администраторам.
func (s *ServiceAdminOrderItem) CreateOrder(ctx context.Context, actor policy.Actor, userID, storeID, groupID int) (domain.OrderCreateDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.OrderCreateDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("store_id", storeID, "group_id", groupID)
	logger.InfoContext(ctx, "creating new order")

	var order domain.OrderCreateDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		// 2. Создание заказа в БД
		order, err = s.order.CreateOrder(ctx, q, userID, storeID, groupID)
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

// GetOrder возвращает заказ по ID вместе с позициями. Доступно только администраторам.
func (s *ServiceAdminOrderItem) GetOrder(ctx context.Context, actor policy.Actor, orderID int) (domain.OrderWithItemDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.OrderWithItemDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("order_id", orderID)
	logger.InfoContext(ctx, "getting order")

	// 2. Получение заказа из БД (без транзакции, только чтение)
	order, err := s.order.GetOrderByID(ctx, s.storage, orderID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get order", "error", err)
		return domain.OrderWithItemDetails{}, fmt.Errorf("get order: %w", err)
	}

	logger.InfoContext(ctx, "order retrieved successfully")
	return order, nil
}

// DeleteOrder удаляет заказ по ID. Доступно только администраторам.
func (s *ServiceAdminOrderItem) DeleteOrder(ctx context.Context, actor policy.Actor, orderID int) error {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("order_id", orderID)
	logger.InfoContext(ctx, "deleting order")

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

// ListOrders возвращает список всех заказов. Доступно только администраторам.
func (s *ServiceAdminOrderItem) ListOrders(ctx context.Context, actor policy.Actor) ([]domain.OrderDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return []domain.OrderDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx)
	logger.InfoContext(ctx, "listing orders")

	// 2. Получение списка заказов из БД (без транзакции)
	orders, err := s.order.ListAdminOrders(ctx, s.storage)
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

// AddItem добавляет позицию в заказ (или обновляет количество, если уже существует).
// Доступно только администраторам.
func (s *ServiceAdminOrderItem) AddItem(ctx context.Context, actor policy.Actor, orderID, productID, unitID, quantity, groupID int) (domain.OrderItemDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.OrderItemDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "product_id", productID, "unit_id", unitID, "quantity", quantity)
	logger.InfoContext(ctx, "adding item to order")

	var item domain.OrderItemDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		// 2. Добавление или обновление позиции
		item, err = s.item.AddItem(ctx, q, orderID, productID, unitID, quantity, groupID)
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

// AddListItems добавляет несколько позиций в заказ (upsert). Доступно только администраторам.
func (s *ServiceAdminOrderItem) AddListItems(ctx context.Context, actor policy.Actor, orderID int, items []domain.OrderItemDetails, groupID int) error {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "count", len(items))
	logger.InfoContext(ctx, "adding list items to order")

	// 2. Upsert каждой позиции в транзакции
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		for _, item := range items {
			// Проверяем, существует ли позиция
			_, err := s.item.GetItemByOrderAndProduct(ctx, q, orderID, item.ProductID)
			if err != nil {
				if domain.IsNotFound(err) {
					// Если нет — добавляем
					if _, err := s.item.AddItem(ctx, q, orderID, item.ProductID, item.UnitID, item.Quantity, groupID); err != nil {
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

// UpdateListItems обновляет список позиций (аналогично AddListItems — upsert). Доступно только администраторам.
func (s *ServiceAdminOrderItem) UpdateListItems(ctx context.Context, actor policy.Actor, orderID int, items []domain.OrderItemDetails, groupID int) error {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "count", len(items))
	logger.InfoContext(ctx, "updating list items")

	// 2. Upsert каждой позиции в транзакции
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		for _, item := range items {
			_, err := s.item.GetItemByOrderAndProduct(ctx, q, orderID, item.ProductID)
			if err != nil {
				if domain.IsNotFound(err) {
					if _, err := s.item.AddItem(ctx, q, orderID, item.ProductID, item.UnitID, item.Quantity, groupID); err != nil {
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

// UpdateItem обновляет количество или единицу измерения позиции. Доступно только администраторам.
func (s *ServiceAdminOrderItem) UpdateItem(ctx context.Context, actor policy.Actor, orderID, productID int, updateOrder domain.OrderItemUpdate) (domain.OrderItemDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.OrderItemDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "product_id", productID)
	logger.InfoContext(ctx, "updating item")

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

// DeleteItem удаляет позицию из заказа. Доступно только администраторам.
func (s *ServiceAdminOrderItem) DeleteItem(ctx context.Context, actor policy.Actor, orderID, productID int) error {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "product_id", productID)
	logger.InfoContext(ctx, "deleting item from order")

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
