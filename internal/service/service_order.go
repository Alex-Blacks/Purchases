package service

import (
	"context"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

func (s *Service) CreateOrder(ctx context.Context, userID, storeID int) (int, error) {
	var orderID int
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		orderID, err = s.order.CreateOrder(ctx, q, userID, storeID)
		return err
	}); err != nil {
		return 0, err
	}

	return orderID, nil
}

func (s *Service) GetOrder(ctx context.Context, userID, orderID int) (domain.OrderWithItems, error) {
	return s.order.GetOrder(ctx, s.storage, userID, orderID)
}

func (s *Service) DeleteOrder(ctx context.Context, userID, orderID int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.order.DeleteOrder(ctx, q, userID, orderID)
	})
}

func (s *Service) ListOrders(ctx context.Context, userID int) ([]domain.Order, error) {
	return s.order.ListOrders(ctx, s.storage, userID)
}

func (s *Service) AddItem(ctx context.Context, orderID, productID, quantity int) (domain.OrderItemDetails, error) {
	var itemID domain.OrderItemDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		itemID, err = s.item.AddItem(ctx, q, orderID, productID, quantity)
		return err
	}); err != nil {
		return itemID, err
	}
	return itemID, nil
}

func (s *Service) UpdateItem(ctx context.Context, orderID, productID, quantity int) (domain.OrderItemDetails, error) {
	var itemID domain.OrderItemDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		itemID, err = s.item.UpdateItem(ctx, q, orderID, productID, quantity)
		return err
	}); err != nil {
		return itemID, err
	}
	return itemID, nil
}

func (s *Service) DeleteItem(ctx context.Context, orderID, productID int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.item.DeleteItem(ctx, q, orderID, productID)
	})
}
