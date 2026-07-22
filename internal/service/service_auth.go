package service

import (
	"context"
	"crypto/sha256"
	"errors"
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

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	logger := logging.LoggerFromContext(ctx).With("email_hash", fmt.Sprintf("%x", sha256.Sum256([]byte(email))))
	ctx = logging.WithContext(ctx, logger)

	user, err := s.svc.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.WarnContext(ctx, "login attempt with non-existent email")
			return "", domain.ErrInvalidCredentials
		}
		logger.ErrorContext(ctx, "database error while fetching user", "error", err)
		return "", fmt.Errorf("get user by email: %w", err)
	}

	logger = logger.With("user_id", user.ID)
	ctx = logging.WithContext(ctx, logger)

	if user.Status != "active" {
		logger.WarnContext(ctx, "login attempt by blocked user")
		return "", domain.ErrInvalidCredentials
	}

	if err := s.svc.CheckPassword(user, password); err != nil {
		logger.WarnContext(ctx, "failed password attempt")
		return "", domain.ErrInvalidCredentials
	}

	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(s.tokenLifetime).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.secret))
	if err != nil {
		logger.ErrorContext(ctx, "failed to sign token", "error", err)
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	logger.InfoContext(ctx, "user logged in")
	return signedToken, nil
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) (string, error) {
	logger := logging.LoggerFromContext(ctx).With("email_hash", fmt.Sprintf("%x", sha256.Sum256([]byte(email))))
	ctx = logging.WithContext(ctx, logger)

	user, err := s.svc.GetUserByEmail(ctx, email)
	if err == nil {
		if user.Status != "active" {
			logger.WarnContext(ctx, "attempt to register a blocked user")
			return "", domain.ErrStatusBlocked
		}
		logger.WarnContext(ctx, "registration attempt with existing email")
		return "", fmt.Errorf("email already registered")
	}

	if !errors.Is(err, domain.ErrNotFound) {
		logger.ErrorContext(ctx, "db error during user check", "error", err)
		return "", fmt.Errorf("check user existence: %w", err)
	}

	user, err = s.svc.CreateUser(ctx, name, password, email, string(policy.RoleUser), "active")
	if err != nil {
		logger.ErrorContext(ctx, "db error during user create", "error", err)
		return "", fmt.Errorf("create user failed: %w", err)
	}

	logger = logger.With("user_id", user.ID)
	ctx = logging.WithContext(ctx, logger)

	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(s.tokenLifetime).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.secret))
	if err != nil {
		logger.ErrorContext(ctx, "failed to sign token", "error", err)
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	logger.InfoContext(ctx, "new user registered")
	return signedToken, nil
}
