package dto

import (
	"fmt"
	"net/mail"
	"strings"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (l *LoginRequest) Validate() error {
	if _, err := mail.ParseAddress(l.Email); err != nil {
		return fmt.Errorf("invalid email format")
	}
	password := strings.TrimSpace(l.Password)
	if password == "" {
		return fmt.Errorf("password must not be empty")
	}
	if len(password) < 8 {
		return fmt.Errorf("password must be more than 8 characters long")
	}

	return nil
}
