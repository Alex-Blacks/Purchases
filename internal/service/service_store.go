package service

import (
	"context"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

func (s *Service) CreateStore(ctx context.Context, name string) (int, error) {
	var storeID int
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		storeID, err = s.store.CreateStore(ctx, q, name)
		return err
	}); err != nil {
		return 0, err
	}

	return storeID, nil
}

func (s *Service) GetStore(ctx context.Context, id int) (domain.Store, error) {
	return s.store.GetStore(ctx, s.storage, id)
}

func (s *Service) DeleteStore(ctx context.Context, id int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.store.DeleteStore(ctx, q, id)
	})
}

func (s *Service) ListStores(ctx context.Context) ([]domain.Store, error) {
	return s.store.ListStores(ctx, s.storage)
}
