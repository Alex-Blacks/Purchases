package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
)

type BaseService struct {
	storage domain.Storage
}

// withTx выполняет функцию в транзакции.
func (s *BaseService) withTx(ctx context.Context, fn func(q domain.Querier) error) (err error) {
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

// resolveGroupID определяет целевой groupID на основе роли и переданного значения
func (b *BaseService) resolveGroupID(actor policy.Actor, groupID *int) (int, error) {
	if actor.HasRole(policy.RoleAdmin) {
		if groupID == nil || *groupID < 1 {
			return 0, domain.ErrGroupIDRequired
		}
		return *groupID, nil
	}
	return actor.GroupID, nil
}

// AccessRead — общая проверка чтения для любой сущности, имеющей GetGroupID()
func (s *BaseService) accessRead(ctx context.Context, q domain.Querier, actor policy.Actor, id int, getter func(ctx context.Context, q domain.Querier, id int) (domain.GroupedEntity, error)) (domain.GroupedEntity, error) {
	logger := logging.LoggerFromContext(ctx).With("entity_id", id)
	logger.InfoContext(ctx, "checking read access")

	// 1. Получение сущности из БД
	entity, err := getter(ctx, q, id)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get entity", "error", err)
		return nil, fmt.Errorf("get entity: %w", err)
	}

	// 2. Проверка прав на чтение через policy
	if err := policy.CanGroupAccessForReading(actor, entity); err != nil {
		logger.WarnContext(ctx, "read access denied", "error", err)
		return nil, err
	}

	logger.InfoContext(ctx, "read access granted")
	return entity, nil
}

// AccessWrite  — общая проверка записи для любой сущности, имеющей GetGroupID()
func (s *BaseService) accessWrite(ctx context.Context, q domain.Querier, actor policy.Actor, id int, getter func(ctx context.Context, q domain.Querier, id int) (domain.GroupedEntity, error)) error {
	logger := logging.LoggerFromContext(ctx).With("entity_id", id)
	logger.InfoContext(ctx, "checking write access")

	// 1. Получение сущности из БД
	entity, err := getter(ctx, q, id)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get entity", "error", err)
		return fmt.Errorf("get entity: %w", err)
	}

	// 2. Проверка прав на изменение через policy
	if err := policy.CanGroupAccessForModify(actor, entity); err != nil {
		logger.WarnContext(ctx, "write access denied", "error", err)
		return err
	}

	logger.InfoContext(ctx, "write access granted")
	return nil
}
