package service

import (
	"context"
	"fmt"

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

func (s *ServiceStore) WithTx(ctx context.Context, fn func(q domain.Querier) error) (err error) {
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

func (s *ServiceStore) CreateStore(ctx context.Context, name string) (domain.Store, error) {
	var store domain.Store
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		store, err = s.store.CreateStore(ctx, q, name)
		return err
	}); err != nil {
		return domain.Store{}, err
	}

	return store, nil
}

func (s *ServiceStore) GetStore(ctx context.Context, id int) (domain.Store, error) {
	return s.store.GetStore(ctx, s.storage, id)
}

func (s *ServiceStore) DeleteStore(ctx context.Context, id int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.store.DeleteStore(ctx, q, id)
	})
}

func (s *ServiceStore) ListStores(ctx context.Context) ([]domain.Store, error) {
	return s.store.ListStores(ctx, s.storage)
}
