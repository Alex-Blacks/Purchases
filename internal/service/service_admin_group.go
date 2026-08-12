package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
)

type ServiceAdminGroup struct {
	storage domain.Storage
	group   domain.GroupRepository
}

func NewServiceAdminGroup(st domain.Storage, group domain.GroupRepository) *ServiceAdminGroup {
	return &ServiceAdminGroup{
		storage: st,
		group:   group,
	}
}

// WithTx выполняет функцию в транзакции.
func (s *ServiceAdminGroup) WithTx(ctx context.Context, fn func(q domain.Querier) error) (err error) {
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

// CreateGroup создаёт новую группу. Доступно только для пользователей с ролью Admin.
func (s *ServiceAdminGroup) CreateGroup(ctx context.Context, actor policy.Actor, name string, adminUserID int) (domain.GroupDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("name", name, "admin_user_id", adminUserID)
	logger.InfoContext(ctx, "creating new group")

	// 1. Проверка прав: только администратор может создавать группы
	if !actor.HasRole(policy.RoleAdmin) {
		logger.WarnContext(ctx, "user is not admin")
		return domain.GroupDetails{}, policy.ErrForbidden
	}

	var group domain.GroupDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		// 2. Создание группы в БД
		group, err = s.group.CreateGroup(ctx, q, name, &adminUserID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to create group", "error", err)
			return fmt.Errorf("create group: %w", err)
		}
		return nil
	}); err != nil {
		return domain.GroupDetails{}, fmt.Errorf("create group: %w", err)
	}

	logger.InfoContext(ctx, "group created successfully", "group_id", group.ID)
	return group, nil
}

// GetGroup возвращает информацию о группе по ID. Доступно только для администраторов.
func (s *ServiceAdminGroup) GetGroup(ctx context.Context, actor policy.Actor, groupID int) (domain.GroupDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("group_id", groupID)
	logger.InfoContext(ctx, "getting group by id")

	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		logger.WarnContext(ctx, "user is not admin")
		return domain.GroupDetails{}, policy.ErrForbidden
	}

	// 2. Получение группы из БД
	group, err := s.group.GetGroupByID(ctx, s.storage, groupID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get group", "error", err)
		return domain.GroupDetails{}, fmt.Errorf("get group: %w", err)
	}

	logger.InfoContext(ctx, "group retrieved successfully")
	return group, nil
}

// UpdateGroup обновляет данные группы. Доступно только для администраторов.
func (s *ServiceAdminGroup) UpdateGroup(ctx context.Context, actor policy.Actor, groupID int, updateGroup domain.GroupUpdate) (domain.GroupDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("group_id", groupID)
	logger.InfoContext(ctx, "updating group")

	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		logger.WarnContext(ctx, "user is not admin")
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

// DeleteGroup удаляет группу по ID. Доступно только для администраторов.
func (s *ServiceAdminGroup) DeleteGroup(ctx context.Context, actor policy.Actor, groupID int) error {
	logger := logging.LoggerFromContext(ctx).With("group_id", groupID)
	logger.InfoContext(ctx, "deleting group")

	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		logger.WarnContext(ctx, "user is not admin")
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

// ListGroups возвращает список всех групп. Доступно только для администраторов.
func (s *ServiceAdminGroup) ListGroups(ctx context.Context, actor policy.Actor) ([]domain.GroupDetails, error) {
	logger := logging.LoggerFromContext(ctx)
	logger.InfoContext(ctx, "listing groups")

	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		logger.WarnContext(ctx, "user is not admin")
		return []domain.GroupDetails{}, policy.ErrForbidden
	}

	// 2. Получение списка групп из БД (без транзакции, только чтение)
	groups, err := s.group.ListGroups(ctx, s.storage)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list groups", "error", err)
		return []domain.GroupDetails{}, fmt.Errorf("list groups: %w", err)
	}

	logger.InfoContext(ctx, "groups listed successfully", "count", len(groups))
	return groups, nil
}
