package service

import (
	"context"
	"fmt"
	"strings"

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

// withTx выполняет функцию в транзакции.
func (s *ServiceStore) withTx(ctx context.Context, fn func(q domain.Querier) error) (err error) {
	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				logging.LoggerFromContext(ctx).ErrorContext(ctx, "rollback failed", "error", rollbackErr)
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
func (s *ServiceStore) accessReadStore(ctx context.Context, q domain.Querier, actor policy.Actor, storeID int) (domain.StoreDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("store_id", storeID)
	logger.InfoContext(ctx, "checking read access for store")

	// 1. Получение магазина из БД
	store, err := s.store.GetStoreByID(ctx, q, storeID)
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

// accessWriteStore проверяет доступ к изменению магазина.
func (s *ServiceStore) accessWriteStore(ctx context.Context, q domain.Querier, actor policy.Actor, storeID int) error {
	logger := logging.LoggerFromContext(ctx).With("store_id", storeID)
	logger.InfoContext(ctx, "checking write access for store")

	// 1. Получение магазина из БД
	store, err := s.store.GetStoreByID(ctx, q, storeID)
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

// CreateStore создаёт новый магазин в указанной группе или группе актора.
func (s *ServiceStore) CreateStore(ctx context.Context, actor policy.Actor, name string, groupID *int) (domain.StoreDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("name", name)
	if groupID != nil {
		logger = logger.With("requested_group_id", groupID)
	}
	logger.InfoContext(ctx, "creating new store")

	if strings.TrimSpace(name) == "" {
		return domain.StoreDetails{}, domain.ErrEmptyName
	}
	var store domain.StoreDetails
	if err := s.withTx(ctx, func(q domain.Querier) error {
		var err error
		if actor.HasRole(policy.RoleAdmin) {
			// Создание магазина с groupID которое передал админ
			if groupID == nil || *groupID < 1 {
				return domain.ErrGroupIDRequired
			}
			store, err = s.store.CreateStore(ctx, q, name, *groupID)
		} else {
			// Создание магазина с groupID актора
			if groupID != nil {
				return domain.ErrGroupIDNotAllowed
			}
			store, err = s.store.CreateStore(ctx, q, name, actor.GroupID)
		}
		if err != nil {
			logger.ErrorContext(ctx, "failed to create store", "error", err)
			return fmt.Errorf("create store: %w", err)
		}
		return nil
	}); err != nil {
		return domain.StoreDetails{}, err
	}

	logger.InfoContext(ctx, "store created successfully", "store_id", store.ID)
	return store, nil
}

// GetStore возвращает магазин по ID с проверкой прав на чтение.
func (s *ServiceStore) GetStore(ctx context.Context, actor policy.Actor, storeID int) (domain.StoreDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("store_id", storeID)
	logger.InfoContext(ctx, "getting store by id")

	if storeID < 1 {
		return domain.StoreDetails{}, domain.ErrInvalidInput
	}
	store, err := s.accessReadStore(ctx, s.storage, actor, storeID)
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

	if storeID < 1 {
		return domain.StoreDetails{}, domain.ErrInvalidInput
	}
	if updateStore.Name != nil && strings.TrimSpace(*updateStore.Name) == "" {
		return domain.StoreDetails{}, domain.ErrEmptyName
	}
	var store domain.StoreDetails
	if err := s.withTx(ctx, func(q domain.Querier) error {
		var err error
		// 1. Проверка прав на запись
		if err := s.accessWriteStore(ctx, q, actor, storeID); err != nil {
			return fmt.Errorf("access write store: %w", err)
		}

		// 2. Обновление магазина в БД
		store, err = s.store.UpdateStoreByID(ctx, q, storeID, updateStore)
		if err != nil {
			logger.ErrorContext(ctx, "failed to update store", "error", err)
			return fmt.Errorf("update store: %w", err)
		}
		return nil
	}); err != nil {
		return domain.StoreDetails{}, err
	}

	logger.InfoContext(ctx, "store updated successfully")
	return store, nil
}

// DeleteStore удаляет магазин с проверкой прав на изменение.
func (s *ServiceStore) DeleteStore(ctx context.Context, actor policy.Actor, storeID int) error {
	logger := logging.LoggerFromContext(ctx).With("store_id", storeID)
	logger.InfoContext(ctx, "deleting store")

	if storeID < 1 {
		return domain.ErrInvalidInput
	}
	if err := s.withTx(ctx, func(q domain.Querier) error {
		// 1. Проверка прав на запись
		if err := s.accessWriteStore(ctx, q, actor, storeID); err != nil {
			return fmt.Errorf("access write store: %w", err)
		}

		// 2. Удаление магазина в БД
		if err := s.store.DeleteStoreByID(ctx, q, storeID); err != nil {
			logger.ErrorContext(ctx, "failed to delete store", "error", err)
			return fmt.Errorf("delete store: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	logger.InfoContext(ctx, "store deleted successfully")
	return nil
}

// ListStores возвращает список магазинов, доступных в группе актора и общей группе.
func (s *ServiceStore) ListStores(ctx context.Context, actor policy.Actor) ([]domain.StoreDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("group_id", actor.GroupID)
	logger.InfoContext(ctx, "listing stores for group and common", "group_id", actor.GroupID, "common_group_id", policy.CommonGroupID)

	// Получение списка магазинов из БД (без транзакции)
	stores, err := s.store.ListStores(ctx, s.storage, []int{actor.GroupID, policy.CommonGroupID})
	if err != nil {
		logger.ErrorContext(ctx, "failed to list stores", "error", err)
		return []domain.StoreDetails{}, fmt.Errorf("list stores: %w", err)
	}

	logger.InfoContext(ctx, "stores listed successfully", "count", len(stores))
	return stores, nil
}

// ListStores возвращает список всех магазинов. Доступно только администраторам.
func (s *ServiceStore) ListAllStores(ctx context.Context, actor policy.Actor) ([]domain.StoreDetails, error) {
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
