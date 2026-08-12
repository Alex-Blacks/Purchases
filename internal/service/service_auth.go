package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	svc           *ServiceUser
	secret        string
	tokenLifetime time.Duration
}

func NewAuthService(svc *ServiceUser, secret string, lifetime time.Duration) *AuthService {
	return &AuthService{
		svc:           svc,
		secret:        secret,
		tokenLifetime: lifetime,
	}
}

// Login аутентифицирует пользователя по email и паролю.
// При успехе возвращает JWT-токен и время его истечения (Unix timestamp).
func (s *AuthService) Login(ctx context.Context, email, password string) (string, int64, error) {
	// Логируем только хеш email для безопасности
	logger := logging.LoggerFromContext(ctx).With("email_hash", fmt.Sprintf("%x", sha256.Sum256([]byte(email))))
	ctx = logging.WithContext(ctx, logger)
	logger.InfoContext(ctx, "user login attempt")

	// 1. Получение пользователя по email
	user, err := s.svc.GetUserByEmail(ctx, email)
	if err != nil {
		if domain.IsNotFound(err) {
			logger.WarnContext(ctx, "login attempt with non-existent email")
			return "", 0, domain.ErrInvalidCredentials
		}
		logger.ErrorContext(ctx, "database error while fetching user", "error", err)
		return "", 0, fmt.Errorf("get user by email: %w", err)
	}

	logger = logger.With("user_id", user.ID)
	ctx = logging.WithContext(ctx, logger)

	// 2. Проверка статуса пользователя
	if user.Status != "active" {
		logger.WarnContext(ctx, "login attempt by blocked user")
		return "", 0, domain.ErrInvalidCredentials
	}

	// 3. Проверка пароля
	if err := s.svc.CheckPassword(user, password); err != nil {
		logger.WarnContext(ctx, "failed password attempt")
		return "", 0, domain.ErrInvalidCredentials
	}

	// 4. Генерация JWT-токена
	exp := time.Now().Add(s.tokenLifetime).Unix()
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"group": user.GroupID,
		"role":  user.Role,
		"exp":   exp,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.secret))
	if err != nil {
		logger.ErrorContext(ctx, "failed to sign token", "error", err)
		return "", 0, fmt.Errorf("sign token: %w", err)
	}

	logger.InfoContext(ctx, "user logged in successfully")
	return signedToken, exp, nil
}

// Register создаёт нового пользователя и возвращает JWT-токен для автоматического входа.
func (s *AuthService) Register(ctx context.Context, name, email, password string) (string, int64, error) {
	logger := logging.LoggerFromContext(ctx).With("email_hash", fmt.Sprintf("%x", sha256.Sum256([]byte(email))))
	ctx = logging.WithContext(ctx, logger)
	logger.InfoContext(ctx, "user registration attempt")

	// 1. Проверка, существует ли пользователь с таким email
	user, err := s.svc.GetUserByEmail(ctx, email)
	if err == nil {
		// Если пользователь заблокирован — возвращаем соответствующую ошибку
		if user.Status != "active" {
			logger.WarnContext(ctx, "attempt to register a blocked user", "user_id", user.ID)
			return "", 0, domain.ErrStatusBlocked
		}
		// Иначе — email уже занят
		logger.WarnContext(ctx, "registration attempt with existing email")
		return "", 0, domain.ErrEmailConflict
	}
	if !domain.IsNotFound(err) {
		logger.ErrorContext(ctx, "database error during user existence check", "error", err)
		return "", 0, fmt.Errorf("check user existence: %w", err)
	}

	// 2. Создание пользователя
	user, err = s.svc.CreateUser(ctx, name, password, email, string(policy.RoleUser), "active")
	if err != nil {
		logger.ErrorContext(ctx, "failed to create user", "error", err)
		return "", 0, fmt.Errorf("create user: %w", err)
	}

	logger = logger.With("user_id", user.ID)
	ctx = logging.WithContext(ctx, logger)

	// 3. Генерация JWT-токена для автоматического входа после регистрации
	exp := time.Now().Add(s.tokenLifetime).Unix()
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"group": user.GroupID,
		"role":  user.Role,
		"exp":   exp,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.secret))
	if err != nil {
		logger.ErrorContext(ctx, "failed to sign token", "error", err)
		return "", 0, fmt.Errorf("sign token: %w", err)
	}

	logger.InfoContext(ctx, "new user registered successfully")
	return signedToken, exp, nil
}
