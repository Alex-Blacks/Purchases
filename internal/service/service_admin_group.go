package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
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

func (s *ServiceAdminGroup) CreateGroup(ctx context.Context, actor policy.Actor, name string, adminUserID int) (domain.GroupDetails, error) {
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.GroupDetails{}, policy.ErrForbidden
	}
	var group domain.GroupDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		group, err = s.group.CreateGroup(ctx, q, name, &adminUserID)
		return err
	}); err != nil {
		return domain.GroupDetails{}, fmt.Errorf("create group: %w", err)
	}
	return group, nil
}

func (s *ServiceAdminGroup) GetGroup(ctx context.Context, actor policy.Actor, groupID int) (domain.GroupDetails, error) {
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.GroupDetails{}, policy.ErrForbidden
	}
	return s.group.GetGroupByID(ctx, s.storage, groupID)
}

func (s *ServiceAdminGroup) UpdateGroup(ctx context.Context, actor policy.Actor, groupID int, updateGroup domain.GroupUpdate) (domain.GroupDetails, error) {
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.GroupDetails{}, policy.ErrForbidden
	}
	var group domain.GroupDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		group, err = s.group.UpdateGroupByID(ctx, q, groupID, updateGroup)
		return err
	}); err != nil {
		return domain.GroupDetails{}, fmt.Errorf("update group: %w", err)
	}
	return group, nil
}

func (s *ServiceAdminGroup) DeleteGroup(ctx context.Context, actor policy.Actor, groupID int) error {
	if !actor.HasRole(policy.RoleAdmin) {
		return policy.ErrForbidden
	}
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.group.DeleteGroupByID(ctx, s.storage, groupID)
	})
}

func (s *ServiceAdminGroup) ListGroups(ctx context.Context, actor policy.Actor) ([]domain.GroupDetails, error) {
	if !actor.HasRole(policy.RoleAdmin) {
		return []domain.GroupDetails{}, policy.ErrForbidden
	}
	return s.group.ListGroups(ctx, s.storage)
}
