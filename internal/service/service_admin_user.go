package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
	"golang.org/x/crypto/bcrypt"
)

type ServiceAdminUser struct {
	storage domain.Storage
	user    domain.UserRepository
	group   domain.GroupRepository
}

func NewServiceAdminUser(st domain.Storage, user domain.UserRepository, group domain.GroupRepository) *ServiceAdminUser {
	return &ServiceAdminUser{
		storage: st,
		user:    user,
		group:   group,
	}
}

// WithTx выполняет функцию в транзакции.
func (s *ServiceAdminUser) WithTx(ctx context.Context, fn func(q domain.Querier) error) (err error) {
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

// GeneratePassword генерирует хеш пароля. Вспомогательный метод.
func (s *ServiceAdminUser) generatePassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", fmt.Errorf("generate password: %w", err)
	}
	return string(bytes), nil
}

// CreateUser создаёт нового пользователя с личной группой. Доступно только администраторам.
func (s *ServiceAdminUser) CreateUser(ctx context.Context, actor policy.Actor, name, password, email, role, status string) (domain.UserDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.UserDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("email", email, "name", name)
	logger.InfoContext(ctx, "creating new user")

	// 2. Проверка, что email не занят (вне транзакции для раннего обнаружения)
	_, err := s.user.GetUserByEmail(ctx, s.storage, email)
	if err == nil {
		logger.WarnContext(ctx, "email already exists")
		return domain.UserDetails{}, domain.ErrEmailConflict
	}
	if !domain.IsNotFound(err) {
		logger.ErrorContext(ctx, "failed to check email existence", "error", err)
		return domain.UserDetails{}, fmt.Errorf("check email: %w", err)
	}

	// 3. Хеширование пароля
	passwordHash, err := s.generatePassword(password)
	if err != nil {
		logger.ErrorContext(ctx, "failed to hash password", "error", err)
		return domain.UserDetails{}, fmt.Errorf("hash password: %w", err)
	}

	var user domain.UserDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		// 4. Создание личной группы
		group, err := s.group.CreateGroup(ctx, q, "Личная группа "+name, nil)
		if err != nil {
			logger.ErrorContext(ctx, "failed to create personal group", "error", err)
			return fmt.Errorf("create personal group: %w", err)
		}

		// 5. Создание пользователя
		user, err = s.user.CreateUser(ctx, q, name, passwordHash, email, group.ID, role, status)
		if err != nil {
			logger.ErrorContext(ctx, "failed to create user", "error", err)
			return fmt.Errorf("create user: %w", err)
		}

		// 6. Назначение пользователя администратором своей группы
		if err = s.group.UpdateGroupAdmin(ctx, q, group.ID, user.ID); err != nil {
			logger.ErrorContext(ctx, "failed to set group admin", "error", err)
			return fmt.Errorf("set group admin: %w", err)
		}
		return nil
	}); err != nil {
		return domain.UserDetails{}, fmt.Errorf("create user: %w", err)
	}

	logger.InfoContext(ctx, "user created successfully", "user_id", user.ID, "group_id", user.GroupID)
	return user, nil
}

// GetUserByID возвращает пользователя по ID. Доступно только администраторам.
func (s *ServiceAdminUser) GetUserByID(ctx context.Context, actor policy.Actor, userID int) (domain.UserDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.UserDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("user_id", userID)
	logger.InfoContext(ctx, "getting user by id")

	// 2. Получение пользователя из БД (без транзакции)
	user, err := s.user.GetUserByID(ctx, s.storage, userID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get user", "error", err)
		return domain.UserDetails{}, fmt.Errorf("get user: %w", err)
	}

	logger.InfoContext(ctx, "user retrieved successfully")
	return user, nil
}

// GetUserByEmail возвращает пользователя по email. Доступно только администраторам.
func (s *ServiceAdminUser) GetUserByEmail(ctx context.Context, actor policy.Actor, email string) (domain.UserDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.UserDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("email", email)
	logger.InfoContext(ctx, "getting user by email")

	// 2. Получение пользователя из БД (без транзакции)
	user, err := s.user.GetUserByEmail(ctx, s.storage, email)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get user by email", "error", err)
		return domain.UserDetails{}, fmt.Errorf("get user by email: %w", err)
	}

	logger.InfoContext(ctx, "user retrieved successfully", "user_id", user.ID)
	return user, nil
}

// UpdateUser обновляет данные пользователя. Доступно только администраторам.
func (s *ServiceAdminUser) UpdateUser(ctx context.Context, actor policy.Actor, userID int, updateUser domain.UserUpdate) (domain.UserDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.UserDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("user_id", userID)
	logger.InfoContext(ctx, "updating user")

	// 2. Подготовка данных для обновления
	var passwordHash *string
	if updateUser.Password != nil {
		hash, err := s.generatePassword(*updateUser.Password)
		if err != nil {
			logger.ErrorContext(ctx, "failed to hash new password", "error", err)
			return domain.UserDetails{}, fmt.Errorf("hash password: %w", err)
		}
		passwordHash = &hash
	}

	// 3. Валидация email (если передан)
	if updateUser.Email != nil {
		if strings.TrimSpace(*updateUser.Email) == "" {
			logger.WarnContext(ctx, "empty email provided")
			return domain.UserDetails{}, domain.ErrInvalidInput
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
		// 4. Обновление пользователя в БД
		user, err = s.user.UpdateUserByID(ctx, q, userID, updateData)
		if err != nil {
			logger.ErrorContext(ctx, "failed to update user", "error", err)
			return fmt.Errorf("update user: %w", err)
		}
		return nil
	}); err != nil {
		return domain.UserDetails{}, fmt.Errorf("update user: %w", err)
	}

	logger.InfoContext(ctx, "user updated successfully")
	return user, nil
}

// DeleteUser удаляет пользователя по ID. Доступно только администраторам.
func (s *ServiceAdminUser) DeleteUser(ctx context.Context, actor policy.Actor, userID int) error {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("user_id", userID)
	logger.InfoContext(ctx, "deleting user")

	// 2. Удаление пользователя в транзакции
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		if err := s.user.DeleteUserByID(ctx, q, userID); err != nil {
			logger.ErrorContext(ctx, "failed to delete user", "error", err)
			return fmt.Errorf("delete user: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	logger.InfoContext(ctx, "user deleted successfully")
	return nil
}

// ListUsers возвращает список всех пользователей. Доступно только администраторам.
func (s *ServiceAdminUser) ListUsers(ctx context.Context, actor policy.Actor) ([]domain.UserDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return []domain.UserDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx)
	logger.InfoContext(ctx, "listing users")

	// 2. Получение списка пользователей из БД (без транзакции)
	users, err := s.user.ListAdminUsers(ctx, s.storage)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list users", "error", err)
		return []domain.UserDetails{}, fmt.Errorf("list users: %w", err)
	}

	logger.InfoContext(ctx, "users listed successfully", "count", len(users))
	return users, nil
}
