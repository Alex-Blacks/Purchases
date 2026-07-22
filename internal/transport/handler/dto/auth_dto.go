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

type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RegisterResponse struct {
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

func (r *RegisterRequest) Validate() error {
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return fmt.Errorf("invalid email format")
	}
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return fmt.Errorf("name must not be empty")
	}
	password := strings.TrimSpace(r.Password)
	if password == "" {
		return fmt.Errorf("password must not be empty")
	}
	if len(password) < 8 {
		return fmt.Errorf("password must be more than 8 characters long")
	}

	return nil
}
