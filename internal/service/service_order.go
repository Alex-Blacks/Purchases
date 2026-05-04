package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

func (s *Service) CreateOrder(ctx context.Context, userID, storeID int) (int, error) {
	var orderID int
	err := s.WithTx(ctx, func(q domain.Queryer) error {
		var err error
		orderID, err = s.order.CreateOrder(ctx, q, userID, storeID)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return orderID, nil
}

func (s *Service) GetOrder(ctx context.Context, userID, orderID int) (domain.OrderWithItemsDTO, error) {
	var result domain.OrderWithItemsDTO
	var order domain.OrderWithItemsDTO
	err := s.WithTx(ctx, func(q domain.Queryer) error {
		var err error
		order, err = s.order.GetOrder(ctx, q, userID, orderID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return result, err
	}

	return order, nil
}

func (s *Service) DeleteOrder(ctx context.Context, userID, orderID int) error {
	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("Error begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			tx.Rollback(ctx)
		}
	}()

	if err := s.order.DeleteOrder(ctx, tx, userID, orderID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return err
		}
		return fmt.Errorf("Error deleted order: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("Error commit transaction: %w", err)
	}

	committed = true
	return nil
}

func (s *Service) ListOrders(ctx context.Context, userID int) ([]domain.OrderDTO, error) {
	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			tx.Rollback(ctx)
		}
	}()

	order, err := s.order.ListOrders(ctx, tx, userID)
	if err != nil {
		return nil, fmt.Errorf("Error get list orders: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("Error commit transaction: %w", err)
	}

	committed = true
	return order, nil
}

func (s *Service) AddItem(ctx context.Context, orderID, productID, qty int) error {
	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("Error begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			tx.Rollback(ctx)
		}
	}()

	if err := s.order.AddItem(ctx, tx, orderID, productID, qty); err != nil {
		return fmt.Errorf("Error add item: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("Error commit transaction: %w", err)
	}

	committed = true
	return nil
}

func (s *Service) UpdateItem(ctx context.Context, orderID, productID, qty int) error {
	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("Error begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			tx.Rollback(ctx)
		}
	}()

	if err := s.order.UpdateItem(ctx, tx, orderID, productID, qty); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return err
		}
		return fmt.Errorf("Error updated item: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("Error commit transaction: %w", err)
	}

	committed = true
	return nil
}

func (s *Service) DeleteItem(ctx context.Context, orderID, productID int) error {
	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("Error begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			tx.Rollback(ctx)
		}
	}()

	if err := s.order.DeleteItem(ctx, tx, orderID, productID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return err
		}
		return fmt.Errorf("Error deleted item: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("Error commit transaction: %w", err)
	}

	committed = true
	return nil
}

func (s *Service) ClearOrder(ctx context.Context, orderID int) error {
	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("Error begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			tx.Rollback(ctx)
		}
	}()

	if err := s.order.ClearOrder(ctx, tx, orderID); err != nil {
		return fmt.Errorf("Error clear list items in order: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("Error commit transaction: %w", err)
	}

	committed = true
	return nil
}
