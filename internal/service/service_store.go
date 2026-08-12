package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
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

// WithTx выполняет функцию в транзакции.
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

// AccessReadStore проверяет доступ к чтению магазина и возвращает его.
func (s *ServiceStore) AccessReadStore(ctx context.Context, actor policy.Actor, storeID int) (domain.StoreDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("store_id", storeID)
	logger.InfoContext(ctx, "checking read access for store")

	// 1. Получение магазина из БД
	store, err := s.store.GetStoreByID(ctx, s.storage, storeID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get store", "error", err)
		return domain.StoreDetails{}, fmt.Errorf("get store: %w", err)
	}

	// 2. Проверка прав на чтение через policy
	if err := policy.CanGroupAccessForReading(actor, store); err != nil {
		logger.WarnContext(ctx, "read access denied", "error", err)
		return domain.StoreDetails{}, err
	}

	logger.InfoContext(ctx, "read access granted")
	return store, nil
}

// AccessWriteStore проверяет доступ к изменению магазина.
func (s *ServiceStore) AccessWriteStore(ctx context.Context, actor policy.Actor, storeID int) error {
	logger := logging.LoggerFromContext(ctx).With("store_id", storeID)
	logger.InfoContext(ctx, "checking write access for store")

	// 1. Получение магазина из БД
	store, err := s.store.GetStoreByID(ctx, s.storage, storeID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get store", "error", err)
		return fmt.Errorf("get store: %w", err)
	}

	// 2. Проверка прав на изменение через policy
	if err := policy.CanGroupAccessForModify(actor, store); err != nil {
		logger.WarnContext(ctx, "write access denied", "error", err)
		return err
	}

	logger.InfoContext(ctx, "write access granted")
	return nil
}

// CreateStore создаёт новый магазин в группе актора.
func (s *ServiceStore) CreateStore(ctx context.Context, actor policy.Actor, name string) (domain.StoreDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("name", name, "group_id", actor.GroupID)
	logger.InfoContext(ctx, "creating new store")

	var store domain.StoreDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		// Создание магазина с groupID актора
		store, err = s.store.CreateStore(ctx, q, name, actor.GroupID)
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

// GetStore возвращает магазин по ID с проверкой прав на чтение.
func (s *ServiceStore) GetStore(ctx context.Context, actor policy.Actor, storeID int) (domain.StoreDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("store_id", storeID)
	logger.InfoContext(ctx, "getting store by id")

	store, err := s.AccessReadStore(ctx, actor, storeID)
	if err != nil {
		return domain.StoreDetails{}, fmt.Errorf("access read store: %w", err)
	}

	logger.InfoContext(ctx, "store retrieved successfully")
	return store, nil
}

// UpdateStore обновляет магазин с проверкой прав на изменение.
func (s *ServiceStore) UpdateStore(ctx context.Context, actor policy.Actor, storeID int, updateStore domain.StoreUpdate) (domain.StoreDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("store_id", storeID)
	logger.InfoContext(ctx, "updating store")

	// 1. Проверка прав на запись
	if err := s.AccessWriteStore(ctx, actor, storeID); err != nil {
		return domain.StoreDetails{}, fmt.Errorf("access write store: %w", err)
	}

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

// DeleteStore удаляет магазин с проверкой прав на изменение.
func (s *ServiceStore) DeleteStore(ctx context.Context, actor policy.Actor, storeID int) error {
	logger := logging.LoggerFromContext(ctx).With("store_id", storeID)
	logger.InfoContext(ctx, "deleting store")

	// 1. Проверка прав на запись
	if err := s.AccessWriteStore(ctx, actor, storeID); err != nil {
		return fmt.Errorf("access write store: %w", err)
	}

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

// ListStores возвращает список магазинов, доступных в группе актора и общей группе.
func (s *ServiceStore) ListStores(ctx context.Context, actor policy.Actor) ([]domain.StoreDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("group_id", actor.GroupID)
	logger.InfoContext(ctx, "listing stores for group")

	// Получение списка магазинов из БД (без транзакции)
	stores, err := s.store.ListStores(ctx, s.storage, []int{actor.GroupID, policy.CommonGroupID})
	if err != nil {
		logger.ErrorContext(ctx, "failed to list stores", "error", err)
		return []domain.StoreDetails{}, fmt.Errorf("list stores: %w", err)
	}

	logger.InfoContext(ctx, "stores listed successfully", "count", len(stores))
	return stores, nil
}
