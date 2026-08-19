package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
)

type ServiceOrderItem struct {
	*BaseService
	orderRepo domain.OrderRepository
	itemRepo  domain.OrderItemRepository
}

func NewServiceOrderItem(st domain.Storage, orderRepo domain.OrderRepository, itemRepo domain.OrderItemRepository) *ServiceOrderItem {
	return &ServiceOrderItem{
		BaseService: &BaseService{storage: st},
		orderRepo:   orderRepo,
		itemRepo:    itemRepo,
	}
}

func (s *ServiceOrderItem) getOrder(ctx context.Context, q domain.Querier, id int) (domain.GroupedEntity, error) {
	return s.orderRepo.GetByID(ctx, q, id)
}

// Create создаёт новый заказ в группе актора.
func (s *ServiceOrderItem) Create(ctx context.Context, actor policy.Actor, storeID int, groupID *int) (domain.OrderCreateDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("store_id", storeID, "group_id", actor.GroupID)
	logger.InfoContext(ctx, "creating new order")

	if storeID < 1 {
		return domain.OrderCreateDetails{}, domain.ErrInvalidInput
	}

	targetGroup, err := s.resolveGroupID(actor, groupID)
	if err != nil {
		return domain.OrderCreateDetails{}, err
	}
	var order domain.OrderCreateDetails
	if err := s.withTx(ctx, func(q domain.Querier) error {
		var err error
		// Создание заказа с userID и groupID актора
		order, err = s.orderRepo.Create(ctx, q, actor.UserID, storeID, targetGroup)
		if err != nil {
			logger.ErrorContext(ctx, "failed to create order", "error", err)
			return fmt.Errorf("create order: %w", err)
		}
		return nil
	}); err != nil {
		return domain.OrderCreateDetails{}, err
	}

	logger.InfoContext(ctx, "order created successfully", "order_id", order.ID)
	return order, nil
}

// GetByID возвращает заказ по ID с проверкой прав на чтение.
func (s *ServiceOrderItem) GetByID(ctx context.Context, actor policy.Actor, orderID int) (domain.OrderWithItemDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID)
	logger.InfoContext(ctx, "getting order")

	if orderID < 1 {
		return domain.OrderWithItemDetails{}, domain.ErrInvalidInput
	}
	order, err := s.accessRead(ctx, s.storage, actor, orderID, s.getOrder)
	if err != nil {
		return domain.OrderWithItemDetails{}, fmt.Errorf("access read order: %w", err)
	}

	logger.InfoContext(ctx, "order retrieved successfully")
	return order.(domain.OrderWithItemDetails), nil
}

// DeleteByID удаляет заказ с проверкой прав на изменение.
func (s *ServiceOrderItem) DeleteByID(ctx context.Context, actor policy.Actor, orderID int) error {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID)
	logger.InfoContext(ctx, "deleting order")

	if orderID < 1 {
		return domain.ErrInvalidInput
	}
	if err := s.withTx(ctx, func(q domain.Querier) error {
		// 1. Проверка прав на запись
		if err := s.accessWrite(ctx, q, actor, orderID, s.getOrder); err != nil {
			return fmt.Errorf("access write order: %w", err)
		}

		// 2. Удаление заказа в транзакции
		if err := s.orderRepo.DeleteByID(ctx, q, orderID); err != nil {
			logger.ErrorContext(ctx, "failed to delete order", "error", err)
			return fmt.Errorf("delete order: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	logger.InfoContext(ctx, "order deleted successfully")
	return nil
}

// List возвращает список заказов пользователя в его группе.
func (s *ServiceOrderItem) List(ctx context.Context, actor policy.Actor) ([]domain.OrderDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("user_id", actor.UserID, "group_id", actor.GroupID)
	logger.InfoContext(ctx, "listing orders")

	orders, err := s.orderRepo.List(ctx, s.storage, actor.UserID, actor.GroupID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list orders", "error", err)
		return []domain.OrderDetails{}, fmt.Errorf("list orders: %w", err)
	}

	logger.InfoContext(ctx, "orders listed successfully", "count", len(orders))
	return orders, nil
}

// ListAll возвращает список всех заказов. Доступно только администраторам.
func (s *ServiceOrderItem) ListAll(ctx context.Context, actor policy.Actor) ([]domain.OrderDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return []domain.OrderDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx)
	logger.InfoContext(ctx, "listing orders")

	// 2. Получение списка заказов из БД (без транзакции)
	orders, err := s.orderRepo.ListAll(ctx, s.storage)
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
func (s *ServiceOrderItem) AddItem(ctx context.Context, actor policy.Actor, orderID, productID, unitID, quantity int, groupID *int) (domain.OrderItemDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "product_id", productID, "unit_id", unitID, "quantity", quantity)
	logger.InfoContext(ctx, "adding item to order")

	if orderID < 1 || productID < 1 || unitID < 1 || quantity < 1 {
		return domain.OrderItemDetails{}, domain.ErrInvalidInput
	}

	targetGroup, err := s.resolveGroupID(actor, groupID)
	if err != nil {
		return domain.OrderItemDetails{}, err
	}

	var item domain.OrderItemDetails
	if err := s.withTx(ctx, func(q domain.Querier) error {
		var err error
		// 1. Проверка прав на запись заказа
		if err := s.accessWrite(ctx, q, actor, orderID, s.getOrder); err != nil {
			return fmt.Errorf("access write order: %w", err)
		}
		// 2. Добавление или обновление позиции
		item, err = s.itemRepo.AddItem(ctx, q, orderID, productID, unitID, quantity, targetGroup)
		if err != nil {
			logger.ErrorContext(ctx, "failed to add item", "error", err)
			return fmt.Errorf("add item: %w", err)
		}
		return nil
	}); err != nil {
		return domain.OrderItemDetails{}, err
	}

	logger.InfoContext(ctx, "item added/updated successfully", "new_quantity", item.Quantity)
	return item, nil
}

// AddListItems добавляет несколько позиций в заказ (upsert). Проверяет права на запись заказа.
func (s *ServiceOrderItem) AddListItems(ctx context.Context, actor policy.Actor, orderID int, items []domain.OrderItemCreate, groupID *int) error {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "count", len(items))
	logger.InfoContext(ctx, "adding list items to order")

	if orderID < 1 {
		return domain.ErrInvalidInput
	}

	targetGroup, err := s.resolveGroupID(actor, groupID)
	if err != nil {
		return err
	}
	if err := s.withTx(ctx, func(q domain.Querier) error {
		// 1. Проверка прав на запись заказа
		if err := s.accessWrite(ctx, q, actor, orderID, s.getOrder); err != nil {
			return fmt.Errorf("access write order: %w", err)
		}
		for _, item := range items {
			if item.ProductID < 1 || item.UnitID < 1 || item.Quantity < 1 {
				return domain.ErrInvalidInput
			}
			// 2. Upsert каждой позиции в транзакции
			_, err := s.itemRepo.GetItemByOrderAndProduct(ctx, q, orderID, item.ProductID)
			if err != nil {
				if domain.IsNotFound(err) {
					// Если позиции нет — добавляем
					if _, err := s.itemRepo.AddItem(ctx, q, orderID, item.ProductID, item.UnitID, item.Quantity, targetGroup); err != nil {
						return fmt.Errorf("add item %d: %w", item.ProductID, err)
					}
					continue
				}
				return fmt.Errorf("get item %d: %w", item.ProductID, err)
			}
			// Если есть — обновляем
			if _, err := s.itemRepo.UpdateItem(ctx, q, orderID, item.ProductID, domain.OrderItemUpdate{UnitID: &item.UnitID, Quantity: &item.Quantity}); err != nil {
				return fmt.Errorf("update item %d: %w", item.ProductID, err)
			}
		}
		return nil
	}); err != nil {
		logger.ErrorContext(ctx, "failed to add list items", "error", err)
		return err
	}

	logger.InfoContext(ctx, "all items upserted successfully")
	return nil
}

// UpdateListItems обновляет список позиций (аналогично AddListItems — upsert). Проверяет права на запись заказа.
func (s *ServiceOrderItem) UpdateListItems(ctx context.Context, actor policy.Actor, orderID int, items []domain.OrderItemCreate, groupID *int) error {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "count", len(items))
	logger.InfoContext(ctx, "updating list items")

	if orderID < 1 {
		return domain.ErrInvalidInput
	}

	targetGroup, err := s.resolveGroupID(actor, groupID)
	if err != nil {
		return err
	}
	if err := s.withTx(ctx, func(q domain.Querier) error {
		// 1. Проверка прав на запись заказа
		if err := s.accessWrite(ctx, q, actor, orderID, s.getOrder); err != nil {
			return fmt.Errorf("access write order: %w", err)
		}
		for _, item := range items {
			if item.ProductID < 1 || item.UnitID < 1 || item.Quantity < 1 {
				return domain.ErrInvalidInput
			}
			// 2. Upsert каждой позиции в транзакции
			_, err := s.itemRepo.GetItemByOrderAndProduct(ctx, q, orderID, item.ProductID)
			if err != nil {
				if domain.IsNotFound(err) {
					if _, err := s.itemRepo.AddItem(ctx, q, orderID, item.ProductID, item.UnitID, item.Quantity, targetGroup); err != nil {
						return fmt.Errorf("add item %d: %w", item.ProductID, err)
					}
					continue
				}
				return fmt.Errorf("get item %d: %w", item.ProductID, err)
			}
			if _, err := s.itemRepo.UpdateItem(ctx, q, orderID, item.ProductID, domain.OrderItemUpdate{UnitID: &item.UnitID, Quantity: &item.Quantity}); err != nil {
				return fmt.Errorf("update item %d: %w", item.ProductID, err)
			}
		}
		return nil
	}); err != nil {
		logger.ErrorContext(ctx, "failed to update list items", "error", err)
		return err
	}

	logger.InfoContext(ctx, "all items upserted successfully")
	return nil
}

// UpdateItem обновляет количество или единицу измерения позиции. Проверяет права на запись заказа.
func (s *ServiceOrderItem) UpdateItem(ctx context.Context, actor policy.Actor, orderID, productID int, updateOrder domain.OrderItemUpdate) (domain.OrderItemDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "product_id", productID)
	logger.InfoContext(ctx, "updating item")

	if orderID < 1 || productID < 1 {
		return domain.OrderItemDetails{}, domain.ErrInvalidInput
	}
	if updateOrder.UnitID != nil && *updateOrder.UnitID < 1 {
		return domain.OrderItemDetails{}, domain.ErrInvalidInput
	}
	if updateOrder.Quantity != nil && *updateOrder.Quantity < 1 {
		return domain.OrderItemDetails{}, domain.ErrInvalidInput
	}
	var item domain.OrderItemDetails
	if err := s.withTx(ctx, func(q domain.Querier) error {
		var err error
		// 1. Проверка прав на запись заказа
		if err := s.accessWrite(ctx, q, actor, orderID, s.getOrder); err != nil {
			return fmt.Errorf("access write order: %w", err)
		}
		// 2. Обновление позиции в БД
		item, err = s.itemRepo.UpdateItem(ctx, q, orderID, productID, updateOrder)
		if err != nil {
			logger.ErrorContext(ctx, "failed to update item", "error", err)
			return fmt.Errorf("update item: %w", err)
		}
		return nil
	}); err != nil {
		return domain.OrderItemDetails{}, err
	}

	logger.InfoContext(ctx, "item updated successfully", "new_quantity", item.Quantity)
	return item, nil
}

// DeleteItem удаляет позицию из заказа. Проверяет права на запись заказа.
func (s *ServiceOrderItem) DeleteItem(ctx context.Context, actor policy.Actor, orderID, productID int) error {
	logger := logging.LoggerFromContext(ctx).With("order_id", orderID, "product_id", productID)
	logger.InfoContext(ctx, "deleting item from order")

	if orderID < 1 || productID < 1 {
		return domain.ErrInvalidInput
	}

	if err := s.withTx(ctx, func(q domain.Querier) error {
		// 1. Проверка прав на запись заказа
		if err := s.accessWrite(ctx, q, actor, orderID, s.getOrder); err != nil {
			return fmt.Errorf("access write order: %w", err)
		}

		// 2. Удаление позиции в транзакции
		if err := s.itemRepo.DeleteItemByOrderAndProduct(ctx, q, orderID, productID); err != nil {
			logger.ErrorContext(ctx, "failed to delete item", "error", err)
			return fmt.Errorf("delete item: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	logger.InfoContext(ctx, "item deleted successfully")
	return nil
}
