package service

import (
	"context"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

func (s *Service) CreateCategory(ctx context.Context, name string) (int, error) {
	var categoryID int
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		categoryID, err = s.category.CreateCategory(ctx, q, name)
		return err
	}); err != nil {
		return 0, err
	}
	return categoryID, nil
}

func (s *Service) GetCategory(ctx context.Context, id int) (domain.Category, error) {
	return s.category.GetCategory(ctx, s.storage, id)
}

func (s *Service) DeleteCategory(ctx context.Context, id int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.category.DeleteCategory(ctx, q, id)
	})
}

func (s *Service) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.category.ListCategories(ctx, s.storage)
}
