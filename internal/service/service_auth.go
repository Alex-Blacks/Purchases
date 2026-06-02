package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Alex-Blacks/Purchases/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	svc    *ServiceUser
	secret string
}

func NewAuthService(svc *ServiceUser, secret string) *AuthService {
	return &AuthService{
		svc:    svc,
		secret: secret,
	}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.svc.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return "", err
		}
		return "", fmt.Errorf("get user by email: %w", err)
	}

	if user.Status != "active" {
		return "", domain.ErrStatusBlocked
	}

	if err := s.svc.CheckPassword(user, password); err != nil {
		return "", domain.ErrIncorrectPassword
	}

	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(1 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signedToken, nil
}
