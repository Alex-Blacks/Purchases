package service

import (
	"context"

	"github.com/Alex-Blacks/Purchases/internal/domain"
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

func (s *ServiceOrderItem) GetAccessibleOrder(ctx context.Context, actor policy.Actor, orderID int) (domain.OrderWithItemDetails, error) {
	order, err := s.order.GetOrder(ctx, s.storage, actor.UserID, orderID)
	if err != nil {
		return domain.OrderWithItemDetails{}, err
	}
	if err := policy.CanAccess(actor, order.Order); err != nil {
		return domain.OrderWithItemDetails{}, err
	}
	return order, nil
}

func (s *ServiceOrderItem) CreateOrder(ctx context.Context, actor policy.Actor, storeID int) (int, error) {
	var orderID int
	if err := s.storage.WithTx(ctx, func(q domain.Querier) error {
		var err error
		orderID, err = s.order.CreateOrder(ctx, q, actor.UserID, storeID)
		return err
	}); err != nil {
		return 0, err
	}

	return orderID, nil
}

func (s *ServiceOrderItem) GetOrder(ctx context.Context, actor policy.Actor, orderID int) (domain.OrderWithItemDetails, error) {
	order, err := s.GetAccessibleOrder(ctx, actor, orderID)
	if err != nil {
		return domain.OrderWithItemDetails{}, err
	}
	return order, nil
}

func (s *ServiceOrderItem) DeleteOrder(ctx context.Context, actor policy.Actor, orderID int) error {
	_, err := s.GetAccessibleOrder(ctx, actor, orderID)
	if err != nil {
		return err
	}
	return s.storage.WithTx(ctx, func(q domain.Querier) error {
		return s.order.DeleteOrder(ctx, q, actor.UserID, orderID)
	})
}

func (s *ServiceOrderItem) ListOrders(ctx context.Context, actor policy.Actor) ([]domain.OrderDetails, error) {
	if err := policy.CanList(actor); err != nil {
		return nil, err
	}
	return s.order.ListOrders(ctx, s.storage, actor.UserID)
}

func (s *ServiceOrderItem) AddItem(ctx context.Context, actor policy.Actor, orderID, productID, quantity int) (domain.OrderItemDetails, error) {
	_, err := s.GetAccessibleOrder(ctx, actor, orderID)
	if err != nil {
		return domain.OrderItemDetails{}, err
	}
	var itemID domain.OrderItemDetails
	if err := s.storage.WithTx(ctx, func(q domain.Querier) error {
		var err error
		itemID, err = s.item.AddItem(ctx, q, orderID, productID, quantity)
		return err
	}); err != nil {
		return itemID, err
	}
	return itemID, nil
}

func (s *ServiceOrderItem) UpdateItem(ctx context.Context, actor policy.Actor, orderID, productID, quantity int) (domain.OrderItemDetails, error) {
	_, err := s.GetAccessibleOrder(ctx, actor, orderID)
	if err != nil {
		return domain.OrderItemDetails{}, err
	}
	var itemID domain.OrderItemDetails
	if err := s.storage.WithTx(ctx, func(q domain.Querier) error {
		var err error
		itemID, err = s.item.UpdateItem(ctx, q, orderID, productID, quantity)
		return err
	}); err != nil {
		return itemID, err
	}
	return itemID, nil
}

func (s *ServiceOrderItem) DeleteItem(ctx context.Context, actor policy.Actor, orderID, productID int) error {
	_, err := s.GetAccessibleOrder(ctx, actor, orderID)
	if err != nil {
		return err
	}
	return s.storage.WithTx(ctx, func(q domain.Querier) error {
		return s.item.DeleteItem(ctx, q, orderID, productID)
	})
}
