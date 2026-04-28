package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

type Service struct {
	store domain.Store
}

func NewService(store domain.Store) *Service {
	return &Service{store: store}
}

func (s *Service) CreateStore(ctx context.Context, name string) error {
	if name == "" {
		return domain.ErrEmptyName
	}
	if err := s.store.CreateStore(ctx, name); err != nil {
		return fmt.Errorf("Error created store: %w", err)
	}

	return nil
}

func (s *Service) GetStoreById(ctx context.Context, id int) (string, error) {
	if id <= 0 {
		return "", domain.ErrInvalidId
	}
	name, err := s.store.GetStoreById(ctx, id)
	if err != nil {
		return "", fmt.Errorf("Error get store: %w", err)
	}

	return name, nil
}

func (s *Service) DeleteStore(ctx context.Context, id int) error {
	if id <= 0 {
		return domain.ErrInvalidId
	}
	if err := s.store.DeleteStore(ctx, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return err
		}
		return fmt.Errorf("Error delete store: %w", err)
	}

	return nil
}

func (s *Service) ListStore(ctx context.Context) ([]domain.ListStore, error) {
	list, err := s.store.ListStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("errors get list stores: %w", err)
	}

	return list, nil
}
