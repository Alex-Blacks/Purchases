package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/policy"
	"golang.org/x/crypto/bcrypt"
)

type ServiceUser struct {
	storage domain.Storage
	user    domain.UserRepository
	group   domain.GroupRepository
}

func NewServiceUser(st domain.Storage, user domain.UserRepository, group domain.GroupRepository) *ServiceUser {
	return &ServiceUser{
		storage: st,
		user:    user,
		group:   group,
	}
}

func (s *ServiceUser) WithTx(ctx context.Context, fn func(q domain.Querier) error) (err error) {
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

func (s *ServiceUser) GetUserByEmail(ctx context.Context, email string) (domain.UserDetails, error) {
	return s.user.GetUserByEmail(ctx, s.storage, email)
}

func (s *ServiceUser) CheckPassword(user domain.UserDetails, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
}

func (s *ServiceUser) GeneratePassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *ServiceUser) CreateUser(ctx context.Context, name, password, email, role, status string) (domain.UserDetails, error) {
	var user domain.UserDetails
	if _, err := s.user.GetUserByEmail(ctx, s.storage, email); err == nil {
		return user, domain.ErrEmailConflict
	}

	password_hash, err := s.GeneratePassword(password)
	if err != nil {
		return user, fmt.Errorf("generate password failed: %w", err)
	}

	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error

		group, err := s.group.CreateGroup(ctx, q, "Личная группа "+name, nil)
		if err != nil {
			return fmt.Errorf("create group: %w", err)
		}

		user, err = s.user.CreateUser(ctx, q, name, password_hash, email, group.ID, role, status)
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		if err = s.group.UpdateGroupAdmin(ctx, q, group.ID, user.ID); err != nil {
			return fmt.Errorf("update group admin: %w", err)
		}
		return nil
	}); err != nil {
		return user, err
	}
	return user, nil
}

func (s *ServiceUser) GetUserByID(ctx context.Context, actor policy.Actor, userID int) (domain.UserDetails, error) {
	user, err := s.user.GetUserByID(ctx, s.storage, userID)
	if err != nil {
		return domain.UserDetails{}, fmt.Errorf("get user: %w", err)
	}
	if err := policy.CanReadUser(actor, user); err != nil {
		return domain.UserDetails{}, policy.ErrForbidden
	}
	return user, nil
}

func (s *ServiceUser) UpdateUser(ctx context.Context, actor policy.Actor, userID int, updateUser domain.UserUpdate) (domain.UserDetails, error) {
	userByID, err := s.user.GetUserByID(ctx, s.storage, userID)
	if err != nil {
		return domain.UserDetails{}, domain.ErrNotFound
	}
	if err := policy.CanUpdateUser(actor, userByID); err != nil {
		return domain.UserDetails{}, policy.ErrForbidden
	}

	var passwordHash *string
	if updateUser.Password != nil {
		hash, err := s.GeneratePassword(*updateUser.Password)
		if err != nil {
			return domain.UserDetails{}, fmt.Errorf("generate password failed: %w", err)
		}
		passwordHash = &hash
	}
	if updateUser.Email != nil {
		if strings.TrimSpace(*updateUser.Email) == "" {
			return domain.UserDetails{}, domain.ErrInvalidInput
		}
		userByEmail, err := s.GetUserByEmail(ctx, *updateUser.Email)
		if err == nil && userByEmail.ID != actor.UserID {
			return domain.UserDetails{}, domain.ErrConflict
		}
	}
	if updateUser.GroupID != nil {
		if !actor.HasRole(policy.RoleAdmin) {
			return domain.UserDetails{}, policy.ErrForbidden
		}
	}
	if updateUser.Role != nil {
		if !actor.HasRole(policy.RoleAdmin) {
			return domain.UserDetails{}, policy.ErrForbidden
		}
	}
	if updateUser.Status != nil {
		if !actor.HasRole(policy.RoleAdmin) {
			return domain.UserDetails{}, policy.ErrForbidden
		}
	}

	updateData := domain.UserUpdate{
		Name:     updateUser.Name,
		Password: passwordHash,
		Email:    updateUser.Email,
		GroupID:  updateUser.GroupID,
		Role:     updateUser.Role,
		Status:   updateUser.Status,
	}

	var user domain.UserDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		user, err = s.user.UpdateUserByID(ctx, q, userID, updateData)
		return err
	}); err != nil {
		return user, err
	}
	return user, nil
}

func (s *ServiceUser) DeleteUser(ctx context.Context, actor policy.Actor, userID int) error {
	user, err := s.user.GetUserByID(ctx, s.storage, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	isGroupAdmin := s.group.CheckGroupAdmin(ctx, s.storage, actor.GroupID, actor.UserID)
	if err := policy.CanDeleteUser(actor, user, isGroupAdmin); err != nil {
		return policy.ErrForbidden
	}
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.user.DeleteUserByID(ctx, q, userID)
	})
}

func (s *ServiceUser) ListUsers(ctx context.Context, actor policy.Actor) ([]domain.UserDetails, error) {
	return s.user.ListUsersInGroup(ctx, s.storage, actor.GroupID)
}
