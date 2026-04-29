package service

import (
	"context"
	"fmt"
)

func (s *Service) CreateOrder(ctx context.Context, UserID, StoreID int) (int, error) {
	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return 0, fmt.Errorf("Error begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	orderID, err := s.order.CreateOrderTx(ctx, tx, UserID, StoreID)
	if err != nil {
		return 0, fmt.Errorf("Error created order: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("Error commit transaction: %w", err)
	}
	return orderID, nil
}
