package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	userSvc       *ServiceUser
	secret        string
	tokenLifetime time.Duration
}

func NewAuthService(userSvc *ServiceUser, secret string, lifetime time.Duration) *AuthService {
	return &AuthService{
		userSvc:       userSvc,
		secret:        secret,
		tokenLifetime: lifetime,
	}
}

// Login аутентифицирует пользователя по email и паролю.
// При успехе возвращает JWT-токен и время его истечения (Unix timestamp).
func (s *AuthService) Login(ctx context.Context, email, password string) (domain.Login, error) {
	// Логируем только хеш email для безопасности
	logger := logging.LoggerFromContext(ctx).With("email_hash", fmt.Sprintf("%x", sha256.Sum256([]byte(email))))
	ctx = logging.WithContext(ctx, logger)
	logger.InfoContext(ctx, "user login attempt")

	if strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return domain.Login{}, domain.ErrEmptyName
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return domain.Login{}, domain.ErrInvalidInput
	}
	// 1. Получение пользователя по email
	user, err := s.userSvc.GetByEmail(ctx, email)
	if err != nil {
		if domain.IsNotFound(err) {
			logger.WarnContext(ctx, "login attempt with non-existent email")
			return domain.Login{}, domain.ErrInvalidCredentials
		}
		logger.ErrorContext(ctx, "database error while fetching user", "error", err)
		return domain.Login{}, fmt.Errorf("get user by email: %w", err)
	}

	logger = logger.With("user_id", user.ID)
	ctx = logging.WithContext(ctx, logger)

	// 2. Проверка статуса пользователя
	if user.Status != domain.UserStatusActive {
		logger.WarnContext(ctx, "login attempt by blocked user")
		return domain.Login{}, domain.ErrInvalidCredentials
	}

	// 3. Проверка пароля
	if err := s.userSvc.checkPassword(user, password); err != nil {
		logger.WarnContext(ctx, "failed password attempt")
		return domain.Login{}, domain.ErrInvalidCredentials
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
		return domain.Login{}, fmt.Errorf("sign token: %w", err)
	}
	result := domain.Login{Token: signedToken, Exp: exp}

	logger.InfoContext(ctx, "user logged in successfully")
	return result, nil
}

// Register создаёт нового пользователя и возвращает JWT-токен для автоматического входа.
func (s *AuthService) Register(ctx context.Context, name, email, password string) (domain.Login, error) {
	logger := logging.LoggerFromContext(ctx).With("email_hash", fmt.Sprintf("%x", sha256.Sum256([]byte(email))))
	ctx = logging.WithContext(ctx, logger)
	logger.InfoContext(ctx, "user registration attempt")

	if strings.TrimSpace(name) == "" || strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return domain.Login{}, domain.ErrEmptyName
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return domain.Login{}, domain.ErrInvalidInput
	}

	// 1. Создание пользователя
	user, err := s.userSvc.Create(ctx, name, password, email, string(policy.RoleUser), domain.UserStatusActive)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create user", "error", err)
		return domain.Login{}, err
	}

	logger = logger.With("user_id", user.ID)
	ctx = logging.WithContext(ctx, logger)

	// 2. Генерация JWT-токена для автоматического входа после регистрации
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
		return domain.Login{}, fmt.Errorf("sign token: %w", err)
	}
	result := domain.Login{Token: signedToken, Exp: exp}

	logger.InfoContext(ctx, "new user registered successfully")
	return result, nil
}
