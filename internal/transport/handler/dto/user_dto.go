package dto

import (
	"github.com/Alex-Blacks/Purchases/internal/domain"
)

// UserRequest используется для создания пользователя.
type UserRequest struct {
	Name     string  `json:"name" validate:"required,min=1,max=50"`
	Password string  `json:"password" validate:"required,min=8,max=100"`
	Email    string  `json:"email" validate:"required,email"`
	Role     *string `json:"role,omitempty" validate:"oneof=admin user"`
}

// UserUpdateRequest используется для обновления пользователя.
type UserUpdateRequest struct {
	Name     *string `json:"name,omitempty" validate:"min=1,max=50"`
	Password *string `json:"password,omitempty" validate:"min=8,max=100"`
	Email    *string `json:"email,omitempty" validate:"email"`
	Role     *string `json:"role,omitempty" validate:"oneof=admin user"`
	Status   *string `json:"status,omitempty" validate:"oneof=active blocked"`
}

// UserResponse возвращает информацию о пользователе.
type UserResponse struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

// ToUserResponse преобразует domain.UserDetails в UserResponse.
func ToUserResponse(user domain.UserDetails) UserResponse {
	return UserResponse{
		ID:     user.ID,
		Name:   user.Name,
		Email:  user.Email,
		Role:   user.Role,
		Status: user.Status,
	}
}

// ToUserUpdateRequest преобразует dto.UserUpdateRequest в domain.UserUpdate.
func ToUserUpdateRequest(up UserUpdateRequest) domain.UserUpdate {
	return domain.UserUpdate{
		Name:     up.Name,
		Password: up.Password,
		Email:    up.Email,
		Role:     up.Role,
		Status:   up.Status,
	}
}

// ToUserListResponse преобразует слайс domain.UserDetails в слайс UserResponse.
func ToUserListResponse(user []domain.UserDetails) []UserResponse {
	resp := make([]UserResponse, len(user))

	for i, it := range user {
		resp[i] = UserResponse{
			ID:     it.ID,
			Name:   it.Name,
			Email:  it.Email,
			Role:   it.Role,
			Status: it.Status,
		}
	}

	return resp
}
