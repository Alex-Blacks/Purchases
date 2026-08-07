package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

type ServiceAdminStore struct {
	storage domain.Storage
	store   domain.StoreRepository
}

func NewServiceAdminStore(st domain.Storage, store domain.StoreRepository) *ServiceAdminStore {
	return &ServiceAdminStore{
		storage: st,
		store:   store,
	}
}

func (s *ServiceAdminStore) WithTx(ctx context.Context, fn func(q domain.Querier) error) (err error) {
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

func (s *ServiceAdminStore) CreateStore(ctx context.Context, name string, groupID int) (domain.StoreDetails, error) {
	var store domain.StoreDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		store, err = s.store.CreateStore(ctx, q, name, groupID)
		return err
	}); err != nil {
		return domain.StoreDetails{}, err
	}
	return store, nil
}

func (s *ServiceAdminStore) GetStore(ctx context.Context, storeID int) (domain.StoreDetails, error) {
	return s.store.GetStoreByID(ctx, s.storage, storeID)
}

func (s *ServiceAdminStore) UpdateStore(ctx context.Context, storeID int, updateStore domain.StoreUpdate) (domain.StoreDetails, error) {
	var store domain.StoreDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		store, err = s.store.UpdateStoreByID(ctx, q, storeID, updateStore)
		return err
	}); err != nil {
		return domain.StoreDetails{}, err
	}

	return store, nil
}

func (s *ServiceAdminStore) DeleteStore(ctx context.Context, storeID int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.store.DeleteStoreByID(ctx, q, storeID)
	})
}

func (s *ServiceAdminStore) ListStores(ctx context.Context) ([]domain.StoreDetails, error) {
	return s.store.ListAdminStores(ctx, s.storage)
}
