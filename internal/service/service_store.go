package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/policy"
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

func (s *ServiceStore) AccessReadStore(ctx context.Context, actor policy.Actor, storeID int) (domain.StoreDetails, error) {
	store, err := s.store.GetStoreByID(ctx, s.storage, storeID)
	if err != nil {
		return domain.StoreDetails{}, err
	}
	if err := policy.CanGroupAccessForReading(actor, store); err != nil {
		return domain.StoreDetails{}, err
	}
	return store, nil
}

func (s *ServiceStore) AccessWriteStore(ctx context.Context, actor policy.Actor, storeID int) error {
	store, err := s.store.GetStoreByID(ctx, s.storage, storeID)
	if err != nil {
		return err
	}
	if err := policy.CanGroupAccessForModify(actor, store); err != nil {
		return err
	}
	return nil
}

func (s *ServiceStore) CreateStore(ctx context.Context, actor policy.Actor, name string) (domain.StoreDetails, error) {
	var store domain.StoreDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		store, err = s.store.CreateStore(ctx, q, name, actor.GroupID)
		return err
	}); err != nil {
		return domain.StoreDetails{}, err
	}
	return store, nil
}

func (s *ServiceStore) GetStore(ctx context.Context, actor policy.Actor, storeID int) (domain.StoreDetails, error) {
	store, err := s.AccessReadStore(ctx, actor, storeID)
	if err != nil {
		return domain.StoreDetails{}, err
	}
	return store, nil
}

func (s *ServiceStore) UpdateStore(ctx context.Context, actor policy.Actor, storeID int, updateStore domain.StoreUpdate) (domain.StoreDetails, error) {
	if err := s.AccessWriteStore(ctx, actor, storeID); err != nil {
		return domain.StoreDetails{}, err
	}
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

func (s *ServiceStore) DeleteStore(ctx context.Context, actor policy.Actor, storeID int) error {
	if err := s.AccessWriteStore(ctx, actor, storeID); err != nil {
		return err
	}
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.store.DeleteStoreByID(ctx, q, storeID)
	})
}

func (s *ServiceStore) ListStores(ctx context.Context, actor policy.Actor) ([]domain.StoreDetails, error) {
	return s.store.ListStores(ctx, s.storage, []int{actor.GroupID, policy.CommonGroupID})
}
