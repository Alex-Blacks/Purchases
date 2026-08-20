package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
)

type ServiceGroup struct {
	*BaseService
	repo domain.GroupRepository
}

func NewServiceGroup(st domain.Storage, repo domain.GroupRepository) *ServiceGroup {
	return &ServiceGroup{
		BaseService: &BaseService{storage: st},
		repo:        repo,
	}
}

// CreateGroup создаёт новую группу. Доступно только для пользователей с ролью Admin.
func (s *ServiceGroup) CreateGroup(ctx context.Context, actor policy.Actor, name string, adminUserID int) (domain.GroupDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("name", name, "admin_user_id", adminUserID)
	logger.InfoContext(ctx, "creating new group")

	// 1. Проверка прав: только администратор может создавать группы
	if !actor.HasRole(policy.RoleAdmin) {
		logger.WarnContext(ctx, "user is not admin")
		return domain.GroupDetails{}, policy.ErrForbidden
	}

	if adminUserID < 1 {
		return domain.GroupDetails{}, domain.ErrInvalidInput
	}
	if strings.TrimSpace(name) == "" {
		return domain.GroupDetails{}, domain.ErrEmptyName
	}

	var group domain.GroupDetails
	if err := s.withTx(ctx, func(q domain.Querier) error {
		var err error
		// 2. Создание группы в БД
		group, err = s.repo.Create(ctx, q, name, &adminUserID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to create group", "error", err)
			return fmt.Errorf("create group: %w", err)
		}
		return nil
	}); err != nil {
		return domain.GroupDetails{}, err
	}

	logger.InfoContext(ctx, "group created successfully", "group_id", group.ID)
	return group, nil
}

// GetByID возвращает группу по ID с проверкой прав на чтение.
func (s *ServiceGroup) GetByID(ctx context.Context, actor policy.Actor, groupID int) (domain.GroupDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("group_id", groupID)
	logger.InfoContext(ctx, "getting group by id")

	if groupID < 1 {
		return domain.GroupDetails{}, domain.ErrInvalidInput
	}

	// 1. Проверка прав на чтение
	if !policy.IsAccessReadGroup(actor, groupID) {
		logger.WarnContext(ctx, "read access denied")
		return domain.GroupDetails{}, policy.ErrForbidden
	}

	// 2. Получение группы из БД (без транзакции)
	group, err := s.repo.GetByID(ctx, s.storage, groupID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get group", "error", err)
		return domain.GroupDetails{}, fmt.Errorf("get group: %w", err)
	}

	logger.InfoContext(ctx, "group retrieved successfully")
	return group, nil
}

// UpdateByID обновляет группу с проверкой прав на запись.
func (s *ServiceGroup) UpdateByID(ctx context.Context, actor policy.Actor, groupID int, updateGroup domain.GroupUpdate) (domain.GroupDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("group_id", groupID)
	logger.InfoContext(ctx, "updating group")

	if groupID < 1 {
		return domain.GroupDetails{}, domain.ErrInvalidInput
	}

	if updateGroup.Name != nil && strings.TrimSpace(*updateGroup.Name) == "" {
		return domain.GroupDetails{}, domain.ErrEmptyName
	}
	if updateGroup.AdminUserID != nil && *updateGroup.AdminUserID < 1 {
		return domain.GroupDetails{}, domain.ErrInvalidInput
	}
	var group domain.GroupDetails
	if err := s.withTx(ctx, func(q domain.Querier) error {
		var err error
		// 1. Проверка прав на запись
		isGroupAdmin := s.repo.CheckGroupAdmin(ctx, q, groupID, actor.UserID)
		if !policy.IsAccessWriteGroup(actor, isGroupAdmin) {
			logger.WarnContext(ctx, "write access denied")
			return policy.ErrForbidden
		}

		// 2. Обновление группы в БД
		group, err = s.repo.UpdateByID(ctx, q, groupID, updateGroup)
		if err != nil {
			logger.ErrorContext(ctx, "failed to update group", "error", err)
			return fmt.Errorf("update group: %w", err)
		}
		return nil
	}); err != nil {
		return domain.GroupDetails{}, err
	}

	logger.InfoContext(ctx, "group updated successfully")
	return group, nil
}

// DeleteByID удаляет группу с проверкой прав на запись.
func (s *ServiceGroup) DeleteByID(ctx context.Context, actor policy.Actor, groupID int) error {
	logger := logging.LoggerFromContext(ctx).With("group_id", groupID)
	logger.InfoContext(ctx, "deleting group")

	if groupID < 1 {
		return domain.ErrInvalidInput
	}
	if err := s.withTx(ctx, func(q domain.Querier) error {
		// 1. Проверка прав на запись
		isGroupAdmin := s.repo.CheckGroupAdmin(ctx, q, groupID, actor.UserID)
		if !policy.IsAccessWriteGroup(actor, isGroupAdmin) {
			logger.WarnContext(ctx, "write access denied")
			return policy.ErrForbidden
		}

		// 2. Удаление группы в транзакции
		if err := s.repo.DeleteByID(ctx, q, groupID); err != nil {
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

// ListAll возвращает список всех групп. Доступно только для администраторов.
func (s *ServiceGroup) ListAll(ctx context.Context, actor policy.Actor) ([]domain.GroupDetails, error) {
	logger := logging.LoggerFromContext(ctx)
	logger.InfoContext(ctx, "listing groups")

	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		logger.WarnContext(ctx, "user is not admin")
		return []domain.GroupDetails{}, policy.ErrForbidden
	}

	// 2. Получение списка групп из БД (без транзакции, только чтение)
	groups, err := s.repo.ListAll(ctx, s.storage)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list groups", "error", err)
		return []domain.GroupDetails{}, fmt.Errorf("list groups: %w", err)
	}

	logger.InfoContext(ctx, "groups listed successfully", "count", len(groups))
	return groups, nil
}
