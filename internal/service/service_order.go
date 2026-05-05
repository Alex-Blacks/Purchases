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

func (s *Service) GetOrder(ctx context.Context, userID, orderID int) (domain.OrderWithItemsDTO, error) {
	return s.order.GetOrder(ctx, s.storage, userID, orderID)
}

func (s *Service) DeleteOrder(ctx context.Context, userID, orderID int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.order.DeleteOrder(ctx, q, userID, orderID)
	})
}

func (s *Service) ListOrders(ctx context.Context, userID int) ([]domain.OrderDTO, error) {
	return s.order.ListOrders(ctx, s.storage, userID)
}

func (s *Service) AddItem(ctx context.Context, orderID, productID, qty int) error {
	if qty <= 0 {
		return domain.ErrInvalidInput
	}
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.item.AddItem(ctx, q, orderID, productID, qty)
	})
}

func (s *Service) UpdateItem(ctx context.Context, orderID, productID, qty int) error {
	if qty <= 0 {
		return domain.ErrInvalidInput
	}
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.item.UpdateItem(ctx, q, orderID, productID, qty)
	})
}

func (s *Service) DeleteItem(ctx context.Context, orderID, productID int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.item.DeleteItem(ctx, q, orderID, productID)
	})
}

func (s *Service) ClearOrder(ctx context.Context, orderID int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.item.ClearOrder(ctx, q, orderID)
	})
}
