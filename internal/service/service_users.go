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
}

func NewServiceUser(st domain.Storage, user domain.UserRepository) *ServiceUser {
	return &ServiceUser{
		storage: st,
		user:    user,
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

func (s *ServiceUser) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	return s.user.GetUserByEmail(ctx, s.storage, email)
}

func (s *ServiceUser) CheckPassword(user domain.User, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
}

func (s *ServiceUser) GeneratePassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *ServiceUser) GetAccessibleUser(ctx context.Context, actor policy.Actor, userID int) (domain.User, error) {
	user, err := s.user.GetUserByID(ctx, s.storage, userID)
	if err != nil {
		return domain.User{}, err
	}
	if err := policy.CanAccess(actor, user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (s *ServiceUser) CreateUser(ctx context.Context, name, password, email, role, status string) (domain.User, error) {
	var user domain.User
	if _, err := s.user.GetUserByEmail(ctx, s.storage, email); err == nil {
		return user, domain.ErrEmailConflict
	}

	password_hash, err := s.GeneratePassword(password)
	if err != nil {
		return user, fmt.Errorf("generate password failed: %w", err)
	}

	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		user, err = s.user.CreateUser(ctx, q, name, password_hash, email, role, status)
		return err
	}); err != nil {
		return user, err
	}
	return user, nil
}

func (s *ServiceUser) GetUserByID(ctx context.Context, actor policy.Actor, userID int) (domain.User, error) {
	user, err := s.GetAccessibleUser(ctx, actor, userID)
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (s *ServiceUser) DeleteUser(ctx context.Context, actor policy.Actor, userID int) error {
	_, err := s.GetAccessibleUser(ctx, actor, userID)
	if err != nil {
		return err
	}
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.user.DeleteUser(ctx, q, userID)
	})
}

func (s *ServiceUser) ListUsers(ctx context.Context, actor policy.Actor) ([]domain.User, error) {
	return s.user.ListUsers(ctx, s.storage)
}

func (s *ServiceUser) UpdateUser(ctx context.Context, actor policy.Actor, userID int, updateUser domain.UpdateUser) (domain.User, error) {
	_, err := s.GetAccessibleUser(ctx, actor, userID)
	if err != nil {
		return domain.User{}, err
	}
	if updateUser.Name == nil && updateUser.Password == nil && updateUser.Email == nil && updateUser.Role == nil && updateUser.Status == nil {
		return domain.User{}, domain.ErrNoFieldsToUpdate
	}
	if (updateUser.Role != nil || updateUser.Status != nil) && !actor.HasRole(policy.RoleAdmin) {
		return domain.User{}, policy.ErrForbidden
	}

	var passwordHash *string
	if updateUser.Password != nil {
		hash, err := s.GeneratePassword(*updateUser.Password)
		if err != nil {
			return domain.User{}, fmt.Errorf("generate password failed: %w", err)
		}
		passwordHash = &hash
	}
	if updateUser.Email != nil {
		if strings.TrimSpace(*updateUser.Email) == "" {
			return domain.User{}, domain.ErrInvalidInput
		}
		user, err := s.GetUserByEmail(ctx, *updateUser.Email)
		if err == nil && user.ID != actor.UserID {
			return domain.User{}, domain.ErrConflict
		}
	}

	updateData := domain.UpdateUser{
		Name:     updateUser.Name,
		Password: passwordHash,
		Email:    updateUser.Email,
		Role:     updateUser.Role,
		Status:   updateUser.Status,
	}

	var user domain.User
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		user, err = s.user.UpdateUser(ctx, q, userID, updateData)
		return err
	}); err != nil {
		return user, err
	}
	return user, nil
}
