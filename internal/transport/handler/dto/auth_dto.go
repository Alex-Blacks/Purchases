package dto

import "github.com/Alex-Blacks/Purchases/internal/domain"

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=50"`
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=1,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=50"`
}

type LoginResponse struct {
	Token string `json:"token"`
	Exp   int64  `json:"exp"`
}

type RegisterResponse struct {
	Token string `json:"token"`
	Exp   int64  `json:"exp"`
}

func ToLoginResponse(data domain.Login) LoginResponse {
	return LoginResponse{
		Token: data.Token,
		Exp:   data.Exp,
	}
}

func ToRegisterResponse(data domain.Login) RegisterResponse {
	return RegisterResponse{
		Token: data.Token,
		Exp:   data.Exp,
	}
}
