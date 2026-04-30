package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

func (s *Service) CreateOrder(ctx context.Context, userID, storeID int) (int, error) {
	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return 0, fmt.Errorf("Error begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	orderID, err := s.order.CreateOrderTx(ctx, tx, userID, storeID)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			return orderID, err
		}
		return 0, fmt.Errorf("Error created order: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("Error commit transaction: %w", err)
	}
	return orderID, nil
}

func (s *Service) GetOrder(ctx context.Context, userID, orderID int) ([]domain.OrderWithItemsDTO, error) {
	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	order, err := s.order.GetOrderTx(ctx, userID, orderID)
	if err != nil {
		return nil, fmt.Errorf("Error get order: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("Error commit transaction: %w", err)
	}

	return order, nil
}
