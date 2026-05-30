package service

import (
	"context"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

type ServiceCategory struct {
	storage  domain.Storage
	category domain.CategoryRepository
}

func NewServiceCategory(st domain.Storage, category domain.CategoryRepository) *ServiceCategory {
	return &ServiceCategory{
		storage:  st,
		category: category,
	}
}

func (s *ServiceCategory) CreateCategory(ctx context.Context, name string) (int, error) {
	var categoryID int
	if err := s.storage.WithTx(ctx, func(q domain.Querier) error {
		var err error
		categoryID, err = s.category.CreateCategory(ctx, q, name)
		return err
	}); err != nil {
		return 0, err
	}
	return categoryID, nil
}

func (s *ServiceCategory) GetCategory(ctx context.Context, id int) (domain.Category, error) {
	return s.category.GetCategory(ctx, s.storage, id)
}

func (s *ServiceCategory) DeleteCategory(ctx context.Context, id int) error {
	return s.storage.WithTx(ctx, func(q domain.Querier) error {
		return s.category.DeleteCategory(ctx, q, id)
	})
}

func (s *ServiceCategory) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.category.ListCategories(ctx, s.storage)
}
