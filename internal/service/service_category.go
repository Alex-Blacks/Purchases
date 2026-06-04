package service

import (
	"context"
	"fmt"

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

func (s *ServiceCategory) WithTx(ctx context.Context, fn func(q domain.Querier) error) (err error) {
	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = fmt.Errorf("tx err: %v, rollback err: %w", err, rollbackErr)
			}
			return
		}

		if commitErr := tx.Commit(ctx); commitErr != nil {
			err = fmt.Errorf("commit err: %w", commitErr)
		}
	}()

	err = fn(tx)
	return err
}

func (s *ServiceCategory) CreateCategory(ctx context.Context, name string) (domain.Category, error) {
	var category domain.Category
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		category, err = s.category.CreateCategory(ctx, q, name)
		return err
	}); err != nil {
		return domain.Category{}, err
	}
	return category, nil
}

func (s *ServiceCategory) GetCategory(ctx context.Context, id int) (domain.Category, error) {
	return s.category.GetCategory(ctx, s.storage, id)
}

func (s *ServiceCategory) DeleteCategory(ctx context.Context, id int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.category.DeleteCategory(ctx, q, id)
	})
}

func (s *ServiceCategory) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.category.ListCategories(ctx, s.storage)
}
