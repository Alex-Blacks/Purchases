package service

import (
	"context"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

type ServiceStore struct {
	storage domain.Storage
	store   domain.StoreRepository
}

func NewServiceStore(st domain.Storage, store domain.StoreRepository) *ServiceStore {
	return &ServiceStore{
		storage: st,
		store:   store,
	}
}

func (s *ServiceStore) CreateStore(ctx context.Context, name string) (int, error) {
	var storeID int
	if err := s.storage.WithTx(ctx, func(q domain.Querier) error {
		var err error
		storeID, err = s.store.CreateStore(ctx, q, name)
		return err
	}); err != nil {
		return 0, err
	}

	return storeID, nil
}

func (s *ServiceStore) GetStore(ctx context.Context, id int) (domain.Store, error) {
	return s.store.GetStore(ctx, s.storage, id)
}

func (s *ServiceStore) DeleteStore(ctx context.Context, id int) error {
	return s.storage.WithTx(ctx, func(q domain.Querier) error {
		return s.store.DeleteStore(ctx, q, id)
	})
}

func (s *ServiceStore) ListStores(ctx context.Context) ([]domain.Store, error) {
	return s.store.ListStores(ctx, s.storage)
}
