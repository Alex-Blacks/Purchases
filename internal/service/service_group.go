package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
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

func (s *ServiceGroup) GetGroup(ctx context.Context, actor policy.Actor, groupID int) (domain.GroupDetails, error) {
	if !policy.IsAccessReadGroup(actor, groupID) {
		return domain.GroupDetails{}, policy.ErrForbidden
	}
	return s.group.GetGroupByID(ctx, s.storage, groupID)
}

func (s *ServiceGroup) UpdateGroup(ctx context.Context, actor policy.Actor, groupID int, updateGroup domain.GroupUpdate) (domain.GroupDetails, error) {
	isGroupAdmin := s.group.CheckGroupAdmin(ctx, s.storage, groupID, actor.UserID)
	if !policy.IsAccessWriteGroup(actor, isGroupAdmin) {
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

func (s *ServiceGroup) DeleteGroup(ctx context.Context, actor policy.Actor, groupID int) error {
	isGroupAdmin := s.group.CheckGroupAdmin(ctx, s.storage, groupID, actor.UserID)
	if !policy.IsAccessWriteGroup(actor, isGroupAdmin) {
		return policy.ErrForbidden
	}
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.group.DeleteGroupByID(ctx, s.storage, groupID)
	})
}
