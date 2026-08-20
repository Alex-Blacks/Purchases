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

type ServiceUser struct {
	*BaseService
	userRepo  domain.UserRepository
	groupRepo domain.GroupRepository
}

func NewServiceUser(st domain.Storage, user domain.UserRepository, group domain.GroupRepository) *ServiceUser {
	return &ServiceUser{
		BaseService: &BaseService{storage: st},
		userRepo:    user,
		groupRepo:   group,
	}
}

// CheckPassword проверяет соответствие пароля хешу.
func (s *ServiceUser) CheckPassword(user domain.UserDetails, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
}

// GeneratePassword генерирует хеш пароля.
func (s *ServiceUser) GeneratePassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", fmt.Errorf("generate password: %w", err)
	}
	return string(bytes), nil
}

// Create регистрирует нового пользователя с личной группой. Доступно без авторизации.
func (s *ServiceUser) Create(ctx context.Context, name, password, email, role, status string) (domain.UserDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("email", email, "name", name)
	logger.InfoContext(ctx, "registering new user")

	if strings.TrimSpace(name) == "" || strings.TrimSpace(password) == "" || strings.TrimSpace(email) == "" || strings.TrimSpace(role) == "" || strings.TrimSpace(status) == "" {
		return domain.UserDetails{}, domain.ErrEmptyName
	}

	if role != string(policy.RoleUser) && role != string(policy.RoleAdmin) {
		return domain.UserDetails{}, domain.ErrInvalidInput
	}

	if status != domain.UserStatusActive && status != domain.UserStatusBlocked {
		return domain.UserDetails{}, domain.ErrInvalidInput
	}

	var user domain.UserDetails
	if err := s.withTx(ctx, func(q domain.Querier) error {
		// 1. Проверка, что email не занят
		_, err := s.userRepo.GetByEmail(ctx, q, email)
		if err == nil {
			logger.WarnContext(ctx, "email already exists")
			return domain.ErrEmailConflict
		}
		if !domain.IsNotFound(err) {
			logger.ErrorContext(ctx, "failed to check email existence", "error", err)
			return fmt.Errorf("check email: %w", err)
		}
		// 2. Хеширование пароля
		passwordHash, err := s.GeneratePassword(password)
		if err != nil {
			logger.ErrorContext(ctx, "failed to hash password", "error", err)
			return fmt.Errorf("hash password: %w", err)
		}
		// 3. Создание личной группы
		group, err := s.groupRepo.Create(ctx, q, "Личная группа "+name, nil)
		if err != nil {
			logger.ErrorContext(ctx, "failed to create personal group", "error", err)
			return fmt.Errorf("create personal group: %w", err)
		}

		// 4. Создание пользователя
		user, err = s.userRepo.Create(ctx, q, name, passwordHash, email, group.ID, role, status)
		if err != nil {
			logger.ErrorContext(ctx, "failed to create user", "error", err)
			return fmt.Errorf("create user: %w", err)
		}

		// 5. Назначение пользователя администратором своей группы
		if err = s.groupRepo.UpdateGroupAdmin(ctx, q, group.ID, user.ID); err != nil {
			logger.ErrorContext(ctx, "failed to set group admin", "error", err)
			return fmt.Errorf("set group admin: %w", err)
		}
		return nil
	}); err != nil {
		return domain.UserDetails{}, err
	}

	logger.InfoContext(ctx, "user registered successfully", "user_id", user.ID, "group_id", user.GroupID)
	return user, nil
}

// GetByID возвращает пользователя по ID с проверкой прав на чтение.
func (s *ServiceUser) GetByID(ctx context.Context, actor policy.Actor, userID int) (domain.UserDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("user_id", userID)
	logger.InfoContext(ctx, "getting user by id")

	if userID < 1 {
		return domain.UserDetails{}, domain.ErrInvalidInput
	}

	// 1. Получение пользователя из БД (без транзакции)
	user, err := s.userRepo.GetByID(ctx, s.storage, userID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get user", "error", err)
		return domain.UserDetails{}, fmt.Errorf("get user: %w", err)
	}

	// 2. Проверка прав на чтение
	if err := policy.CanReadUser(actor, user); err != nil {
		logger.WarnContext(ctx, "user not allowed to read this user")
		return domain.UserDetails{}, policy.ErrForbidden
	}

	logger.InfoContext(ctx, "user retrieved successfully")
	return user, nil
}

// GetByEmail возвращает пользователя по email. Используется внутри сервиса.
func (s *ServiceUser) GetByEmail(ctx context.Context, email string) (domain.UserDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("email", email)
	logger.InfoContext(ctx, "getting user by email")

	if strings.TrimSpace(email) == "" {
		return domain.UserDetails{}, domain.ErrEmptyName
	}
	user, err := s.userRepo.GetByEmail(ctx, s.storage, email)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get user by email", "error", err)
		return domain.UserDetails{}, fmt.Errorf("get user by email: %w", err)
	}

	logger.InfoContext(ctx, "user retrieved by email", "user_id", user.ID)
	return user, nil
}

// UpdateByID обновляет данные пользователя с проверкой прав.
func (s *ServiceUser) UpdateByID(ctx context.Context, actor policy.Actor, userID int, updateUser domain.UserUpdate) (domain.UserDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("user_id", userID)
	logger.InfoContext(ctx, "updating user")

	// 1. Подготовка данных для обновления
	var passwordHash *string
	if updateUser.Password != nil {
		hash, err := s.GeneratePassword(*updateUser.Password)
		if err != nil {
			logger.ErrorContext(ctx, "failed to hash new password", "error", err)
			return domain.UserDetails{}, fmt.Errorf("hash password: %w", err)
		}
		passwordHash = &hash
	}

	if updateUser.Name != nil && strings.TrimSpace(*updateUser.Name) == "" {
		return domain.UserDetails{}, domain.ErrEmptyName
	}

	// 2. Только администратор может менять GroupID, Role, Status
	if updateUser.GroupID != nil && !actor.HasRole(policy.RoleAdmin) {
		logger.WarnContext(ctx, "attempt to change group without admin role")
		return domain.UserDetails{}, policy.ErrForbidden
	}
	if updateUser.Role != nil && !actor.HasRole(policy.RoleAdmin) {
		logger.WarnContext(ctx, "attempt to change role without admin role")
		return domain.UserDetails{}, policy.ErrForbidden
	}
	if updateUser.Status != nil && !actor.HasRole(policy.RoleAdmin) {
		logger.WarnContext(ctx, "attempt to change status without admin role")
		return domain.UserDetails{}, policy.ErrForbidden
	}

	if updateUser.GroupID != nil && *updateUser.GroupID < 1 {
		return domain.UserDetails{}, domain.ErrInvalidInput
	}

	if updateUser.Role != nil && strings.TrimSpace(*updateUser.Role) == "" {
		return domain.UserDetails{}, domain.ErrEmptyName
	}

	if updateUser.Status != nil && strings.TrimSpace(*updateUser.Status) == "" {
		return domain.UserDetails{}, domain.ErrEmptyName
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
	err := s.withTx(ctx, func(q domain.Querier) error {
		if updateUser.GroupID != nil && *updateUser.GroupID > 0 {
			_, err := s.groupRepo.GetByID(ctx, q, *updateUser.GroupID)
			if err != nil {
				return fmt.Errorf("group not found: %w", err)
			}
		}
		// 3. Проверка существования пользователя и прав на обновление
		userByID, err := s.userRepo.GetByID(ctx, q, userID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to get user for update", "error", err)
			return fmt.Errorf("get user: %w", err)
		}
		if err := policy.CanUpdateUser(actor, userByID); err != nil {
			logger.WarnContext(ctx, "user not allowed to update this user")
			return policy.ErrForbidden
		}
		// 4. Валидация и проверка email на конфликт (если передан)
		if updateUser.Email != nil {
			if strings.TrimSpace(*updateUser.Email) == "" {
				logger.WarnContext(ctx, "empty email provided")
				return domain.ErrInvalidInput
			}
			userByEmail, err := s.userRepo.GetByEmail(ctx, q, *updateUser.Email)
			if err != nil && !domain.IsNotFound(err) {
				logger.ErrorContext(ctx, "failed to check email conflict", "error", err)
				return fmt.Errorf("check email: %w", err)
			}
			if err == nil && userByEmail.ID != userID {
				logger.WarnContext(ctx, "email already used by another user")
				return domain.ErrConflict
			}
		}
		// 5. Обновление пользователя в БД
		user, err = s.userRepo.UpdateByID(ctx, q, userID, updateData)
		if err != nil {
			logger.ErrorContext(ctx, "failed to update user", "error", err)
			return fmt.Errorf("update user: %w", err)
		}
		return nil
	})
	if err != nil {
		return domain.UserDetails{}, err
	}

	logger.InfoContext(ctx, "user updated successfully")
	return user, nil
}

// DeleteByID удаляет пользователя с проверкой прав.
func (s *ServiceUser) DeleteByID(ctx context.Context, actor policy.Actor, userID int) error {
	logger := logging.LoggerFromContext(ctx).With("user_id", userID)
	logger.InfoContext(ctx, "deleting user")

	if userID < 1 {
		return domain.ErrInvalidInput
	}

	if err := s.withTx(ctx, func(q domain.Querier) error {
		// 1. Получение данных пользователя для проверки прав
		user, err := s.userRepo.GetByID(ctx, q, userID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to get user for delete", "error", err)
			return fmt.Errorf("get user: %w", err)
		}

		// 2. Проверка, является ли пользователь администратором своей группы (для политики удаления)
		isGroupAdmin := s.groupRepo.CheckGroupAdmin(ctx, q, actor.GroupID, actor.UserID)
		if err := policy.CanDeleteUser(actor, user, isGroupAdmin); err != nil {
			logger.WarnContext(ctx, "user not allowed to delete this user")
			return policy.ErrForbidden
		}

		// 3. Удаление пользователя в транзакции
		if err := s.userRepo.DeleteByID(ctx, q, userID); err != nil {
			logger.ErrorContext(ctx, "failed to delete user", "error", err)
			return fmt.Errorf("delete user: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	logger.InfoContext(ctx, "user deleted successfully")
	return nil
}

// List возвращает список пользователей в группе актора.
func (s *ServiceUser) List(ctx context.Context, actor policy.Actor) ([]domain.UserDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("group_id", actor.GroupID)
	logger.InfoContext(ctx, "listing users in group")

	// 1. Получение списка пользователей группы из БД (без транзакции)
	users, err := s.userRepo.ListInGroup(ctx, s.storage, actor.GroupID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list users in group", "error", err)
		return []domain.UserDetails{}, fmt.Errorf("list users: %w", err)
	}

	logger.InfoContext(ctx, "users in group listed successfully", "count", len(users))
	return users, nil
}

// ListAll возвращает список всех пользователей. Доступно только администраторам.
func (s *ServiceUser) ListAll(ctx context.Context, actor policy.Actor) ([]domain.UserDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return []domain.UserDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx)
	logger.InfoContext(ctx, "listing users")

	// 2. Получение списка пользователей из БД (без транзакции)
	users, err := s.userRepo.ListAll(ctx, s.storage)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list users", "error", err)
		return []domain.UserDetails{}, fmt.Errorf("list users: %w", err)
	}

	logger.InfoContext(ctx, "users listed successfully", "count", len(users))
	return users, nil
}
