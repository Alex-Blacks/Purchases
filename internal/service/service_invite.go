package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/policy"
	"github.com/google/uuid"
)

type ServiceInvite struct {
	storage domain.Storage
	invite  domain.InviteRepository
	group   domain.GroupRepository
	user    domain.UserRepository
}

func NewServiceInvite(st domain.Storage, invite domain.InviteRepository, group domain.GroupRepository, user domain.UserRepository) *ServiceInvite {
	return &ServiceInvite{
		storage: st,
		invite:  invite,
		group:   group,
		user:    user,
	}
}

func (s *ServiceInvite) WithTx(ctx context.Context, fn func(q domain.Querier) error) (err error) {
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

// CreateInvite создаёт приглашение для пользователя с email inviteeEmail в группу actor.GroupID.
func (s *ServiceInvite) CreateInvite(ctx context.Context, actor policy.Actor, inviteeEmail string) (domain.InviteDetails, error) {
	var invite domain.InviteDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {

		//
		isGroupAdmin, err := s.group.CheckGroupAdmin(ctx, q, actor.GroupID, actor.UserID)
		if err != nil {
			return fmt.Errorf("check group: %w", err)
		}
		if !isGroupAdmin {
			return policy.ErrForbidden
		}

		user, err := s.user.GetUserByEmail(ctx, q, inviteeEmail)
		if err != nil {
			if !errors.Is(err, domain.ErrNotFound) {
				return fmt.Errorf("get user by email: %w", err)
			}
		}

		isGroupAdmin, err = s.group.CheckGroupAdmin(ctx, q, user.GroupID, user.ID)
		if err != nil {
			return fmt.Errorf("check group: %w", err)
		}
		if isGroupAdmin {
			return domain.ErrUserAlreadyInGroup
		}

		invate, err := s.invite.GetInviteByEmail(ctx, q, inviteeEmail, actor.GroupID)
		if err != nil {
			if !errors.Is(err, domain.ErrNotFound) {
				return fmt.Errorf("get invite by email: %w", err)
			}
		}
		if invate.Status == "pending" {
			return fmt.Errorf("the invitation already exists")
		}

		token := uuid.New().String()
		invite, err = s.invite.CreateInvite(ctx, q, actor.GroupID, actor.UserID, inviteeEmail, token)
		return err
	}); err != nil {
		return domain.InviteDetails{}, fmt.Errorf("create invite: %w", err)
	}
	return invite, nil
}
