package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
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

// WithTx выполняет функцию в транзакции.
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

// CreateStore создаёт новый магазин. Доступно только администраторам.
func (s *ServiceAdminStore) CreateStore(ctx context.Context, actor policy.Actor, name string, groupID int) (domain.StoreDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.StoreDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("name", name, "group_id", groupID)
	logger.InfoContext(ctx, "creating new store")

	var store domain.StoreDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		// 2. Создание магазина в БД
		store, err = s.store.CreateStore(ctx, q, name, groupID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to create store", "error", err)
			return fmt.Errorf("create store: %w", err)
		}
		return nil
	}); err != nil {
		return domain.StoreDetails{}, fmt.Errorf("create store: %w", err)
	}

	logger.InfoContext(ctx, "store created successfully", "store_id", store.ID)
	return store, nil
}

// GetStore возвращает магазин по ID. Доступно только администраторам.
func (s *ServiceAdminStore) GetStore(ctx context.Context, actor policy.Actor, storeID int) (domain.StoreDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.StoreDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("store_id", storeID)
	logger.InfoContext(ctx, "getting store by id")

	// 2. Получение магазина из БД (без транзакции)
	store, err := s.store.GetStoreByID(ctx, s.storage, storeID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get store", "error", err)
		return domain.StoreDetails{}, fmt.Errorf("get store: %w", err)
	}

	logger.InfoContext(ctx, "store retrieved successfully")
	return store, nil
}

// UpdateStore обновляет магазин. Доступно только администраторам.
func (s *ServiceAdminStore) UpdateStore(ctx context.Context, actor policy.Actor, storeID int, updateStore domain.StoreUpdate) (domain.StoreDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.StoreDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("store_id", storeID)
	logger.InfoContext(ctx, "updating store")

	var store domain.StoreDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		// 2. Обновление магазина в БД
		store, err = s.store.UpdateStoreByID(ctx, q, storeID, updateStore)
		if err != nil {
			logger.ErrorContext(ctx, "failed to update store", "error", err)
			return fmt.Errorf("update store: %w", err)
		}
		return nil
	}); err != nil {
		return domain.StoreDetails{}, fmt.Errorf("update store: %w", err)
	}

	logger.InfoContext(ctx, "store updated successfully")
	return store, nil
}

// DeleteStore удаляет магазин по ID. Доступно только администраторам.
func (s *ServiceAdminStore) DeleteStore(ctx context.Context, actor policy.Actor, storeID int) error {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("store_id", storeID)
	logger.InfoContext(ctx, "deleting store")

	// 2. Удаление магазина в транзакции
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		if err := s.store.DeleteStoreByID(ctx, q, storeID); err != nil {
			logger.ErrorContext(ctx, "failed to delete store", "error", err)
			return fmt.Errorf("delete store: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("delete store: %w", err)
	}

	logger.InfoContext(ctx, "store deleted successfully")
	return nil
}

// ListStores возвращает список всех магазинов. Доступно только администраторам.
func (s *ServiceAdminStore) ListStores(ctx context.Context, actor policy.Actor) ([]domain.StoreDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return []domain.StoreDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx)
	logger.InfoContext(ctx, "listing stores")

	// 2. Получение списка магазинов из БД (без транзакции)
	stores, err := s.store.ListAdminStores(ctx, s.storage)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list stores", "error", err)
		return []domain.StoreDetails{}, fmt.Errorf("list stores: %w", err)
	}

	logger.InfoContext(ctx, "stores listed successfully", "count", len(stores))
	return stores, nil
}
