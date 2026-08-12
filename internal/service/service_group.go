package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
)

type ServiceGroup struct {
	storage domain.Storage
	group   domain.GroupRepository
}

func NewServiceGroup(st domain.Storage, group domain.GroupRepository) *ServiceGroup {
	return &ServiceGroup{
		storage: st,
		group:   group,
	}
}

// WithTx выполняет функцию в транзакции.
func (s *ServiceGroup) WithTx(ctx context.Context, fn func(q domain.Querier) error) (err error) {
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

// GetGroup возвращает группу по ID с проверкой прав на чтение.
func (s *ServiceGroup) GetGroup(ctx context.Context, actor policy.Actor, groupID int) (domain.GroupDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("group_id", groupID)
	logger.InfoContext(ctx, "getting group by id")

	// 1. Проверка прав на чтение
	if !policy.IsAccessReadGroup(actor, groupID) {
		logger.WarnContext(ctx, "read access denied")
		return domain.GroupDetails{}, policy.ErrForbidden
	}

	// 2. Получение группы из БД (без транзакции)
	group, err := s.group.GetGroupByID(ctx, s.storage, groupID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get group", "error", err)
		return domain.GroupDetails{}, fmt.Errorf("get group: %w", err)
	}

	logger.InfoContext(ctx, "group retrieved successfully")
	return group, nil
}

// UpdateGroup обновляет группу с проверкой прав на запись.
func (s *ServiceGroup) UpdateGroup(ctx context.Context, actor policy.Actor, groupID int, updateGroup domain.GroupUpdate) (domain.GroupDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("group_id", groupID)
	logger.InfoContext(ctx, "updating group")

	// 1. Проверка прав на запись
	isGroupAdmin := s.group.CheckGroupAdmin(ctx, s.storage, groupID, actor.UserID)
	if !policy.IsAccessWriteGroup(actor, isGroupAdmin) {
		logger.WarnContext(ctx, "write access denied")
		return domain.GroupDetails{}, policy.ErrForbidden
	}

	var group domain.GroupDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		// 2. Обновление группы в БД
		group, err = s.group.UpdateGroupByID(ctx, q, groupID, updateGroup)
		if err != nil {
			logger.ErrorContext(ctx, "failed to update group", "error", err)
			return fmt.Errorf("update group: %w", err)
		}
		return nil
	}); err != nil {
		return domain.GroupDetails{}, fmt.Errorf("update group: %w", err)
	}

	logger.InfoContext(ctx, "group updated successfully")
	return group, nil
}

// DeleteGroup удаляет группу с проверкой прав на запись.
func (s *ServiceGroup) DeleteGroup(ctx context.Context, actor policy.Actor, groupID int) error {
	logger := logging.LoggerFromContext(ctx).With("group_id", groupID)
	logger.InfoContext(ctx, "deleting group")

	// 1. Проверка прав на запись
	isGroupAdmin := s.group.CheckGroupAdmin(ctx, s.storage, groupID, actor.UserID)
	if !policy.IsAccessWriteGroup(actor, isGroupAdmin) {
		logger.WarnContext(ctx, "write access denied")
		return policy.ErrForbidden
	}

	// 2. Удаление группы в транзакции
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		if err := s.group.DeleteGroupByID(ctx, q, groupID); err != nil {
			logger.ErrorContext(ctx, "failed to delete group", "error", err)
			return fmt.Errorf("delete group: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("delete group: %w", err)
	}

	logger.InfoContext(ctx, "group deleted successfully")
	return nil
}
