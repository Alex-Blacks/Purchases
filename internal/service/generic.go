package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
)

// GenericService предоставляет общие CRUDL-операции для любой сущности T,
// используя репозиторий R, который реализует GenericRepository[T].
type GenericService[T domain.GroupedEntity, R domain.GenericRepository[T]] struct {
	*BaseService
	repo R
}

func (s *GenericService[T, R]) getEntity(ctx context.Context, q domain.Querier, id int) (domain.GroupedEntity, error) {
	return s.repo.GetByID(ctx, q, id)
}

// Create создаёт новую сущность в указанной группе или группе актора.
func (s *GenericService[T, R]) Create(ctx context.Context, actor policy.Actor, params any, groupID *int) (T, error) {
	var zero T
	logger := logging.LoggerFromContext(ctx).With("create_entity", params)
	logger.InfoContext(ctx, "creating new entity")

	targetGroup, err := s.resolveGroupID(actor, groupID)
	if err != nil {
		return zero, err
	}
	var entity T
	if err := s.withTx(ctx, func(q domain.Querier) error {
		var err error
		entity, err = s.repo.Create(ctx, q, params, targetGroup)
		if err != nil {
			logger.ErrorContext(ctx, "failed to create entity", "error", err)
			return fmt.Errorf("create entity: %w", err)
		}
		return nil
	}); err != nil {
		return zero, err
	}

	logger.InfoContext(ctx, "entity created successfully")
	return entity, nil
}

// Get возвращает сущность по ID с проверкой прав на чтение
func (s *GenericService[T, R]) Get(ctx context.Context, actor policy.Actor, id int) (T, error) {
	var zero T
	logger := logging.LoggerFromContext(ctx).With("entity_id", id)
	logger.InfoContext(ctx, "getting entity by id")

	if id < 1 {
		return zero, domain.ErrInvalidInput
	}

	// Используем AccessRead с функцией получения из репозитория
	entity, err := s.accessRead(ctx, s.storage, actor, id, s.getEntity)
	if err != nil {
		return zero, fmt.Errorf("access read entity: %w", err)
	}

	// Приведение типа к T (безопасно, т.к. репозиторий возвращает именно T)
	result, ok := entity.(T)
	if !ok {
		return zero, fmt.Errorf("unexpected entity type")
	}

	logger.InfoContext(ctx, "entity retrieved successfully")
	return result, nil
}

// Update обновляет сущность с проверкой прав на изменение.
func (s *GenericService[T, R]) Update(ctx context.Context, actor policy.Actor, id int, updates any) (T, error) {
	var zero T
	logger := logging.LoggerFromContext(ctx).With("entity_id", id, "update_entity", updates)
	logger.InfoContext(ctx, "updating entity")

	if id < 1 {
		return zero, domain.ErrInvalidInput
	}
	var entity T
	if err := s.withTx(ctx, func(q domain.Querier) error {
		var err error
		// 1. Проверка прав на запись
		if err := s.accessWrite(ctx, q, actor, id, s.getEntity); err != nil {
			return fmt.Errorf("access write entity: %w", err)
		}

		// 2. Обновление сущности в БД
		entity, err = s.repo.UpdateByID(ctx, q, id, updates)
		if err != nil {
			logger.ErrorContext(ctx, "failed to update entity", "error", err)
			return fmt.Errorf("update entity: %w", err)
		}
		return nil
	}); err != nil {
		return zero, err
	}

	logger.InfoContext(ctx, "entity updated successfully")
	return entity, nil
}

// Delete удаляет сущность с проверкой прав на изменение.
func (s *GenericService[T, R]) Delete(ctx context.Context, actor policy.Actor, id int) error {
	logger := logging.LoggerFromContext(ctx).With("entity_id", id)
	logger.InfoContext(ctx, "deleting ")

	if id < 1 {
		return domain.ErrInvalidInput
	}
	if err := s.withTx(ctx, func(q domain.Querier) error {
		// 1. Проверка прав на запись
		if err := s.accessWrite(ctx, q, actor, id, s.getEntity); err != nil {
			return fmt.Errorf("access write entity: %w", err)
		}

		// 2. Удаление сущности в БД
		if err := s.repo.DeleteByID(ctx, q, id); err != nil {
			logger.ErrorContext(ctx, "failed to delete entity", "error", err)
			return fmt.Errorf("delete entity: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	logger.InfoContext(ctx, "entity deleted successfully")
	return nil
}

// Lists возвращает список сущностей, доступных в группе актора и общей группе.
func (s *GenericService[T, R]) List(ctx context.Context, actor policy.Actor) ([]T, error) {
	logger := logging.LoggerFromContext(ctx).With("group_id", actor.GroupID)
	logger.InfoContext(ctx, "listing s for group and common", "group_id", actor.GroupID, "common_group_id", policy.CommonGroupID)

	// Получение списка сущностей из БД (без транзакции)
	entities, err := s.repo.List(ctx, s.storage, []int{actor.GroupID, policy.CommonGroupID})
	if err != nil {
		logger.ErrorContext(ctx, "failed to list entities", "error", err)
		var zero []T
		return zero, fmt.Errorf("list entities: %w", err)
	}

	logger.InfoContext(ctx, "entities listed successfully", "count", len(entities))
	return entities, nil
}

// Lists возвращает список всех сущностей. Доступно только администраторам.
func (s *GenericService[T, R]) ListAll(ctx context.Context, actor policy.Actor) ([]T, error) {
	var zero []T
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return zero, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx)
	logger.InfoContext(ctx, "listing entities")

	// 2. Получение списка сущностей из БД (без транзакции)
	entities, err := s.repo.ListAll(ctx, s.storage)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list entities", "error", err)
		return zero, fmt.Errorf("list entities: %w", err)
	}

	logger.InfoContext(ctx, "entities listed successfully", "count", len(entities))
	return entities, nil
}
