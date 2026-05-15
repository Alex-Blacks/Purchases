package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func (s *Service) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	return s.user.GetUserByEmail(ctx, s.storage, email)
}

func (s *Service) CheckPassword(user domain.User, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
}
func (s *Service) GeneratePassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
func (s *Service) CreateUser(ctx context.Context, name, password, email, role, status string) (domain.User, error) {
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

func (s *Service) GetUserByID(ctx context.Context, userID int) (domain.User, error) {
	return s.user.GetUserByID(ctx, s.storage, userID)
}
func (s *Service) DeleteUser(ctx context.Context, userID int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.user.DeleteUser(ctx, q, userID)
	})
}

func (s *Service) ListUsers(ctx context.Context) ([]domain.User, error) {
	return s.user.ListUsers(ctx, s.storage)
}
func (s *Service) UpdateUser(ctx context.Context, userID int, updateUser domain.UpdateUser) (domain.User, error) {
	var user domain.User
	if !hasUpdates(updateUser) {
		return user, domain.ErrEmptyUpdate
	}
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		user, err = s.user.UpdateUser(ctx, q, userID, updateUser)
		return err
	}); err != nil {
		return user, err
	}
	return user, nil
}
