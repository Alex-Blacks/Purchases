package dto

import "github.com/Alex-Blacks/Purchases/internal/domain"

type UserRequest struct {
	Name     string  `json:"name" validate:"required"`
	Password string  `json:"password" validate:"required"`
	Email    string  `json:"email" validate:"required"`
	Role     *string `json:"role,omitempty"`
}

type UserResponse struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

func ToUserResponse(user domain.User) UserResponse {
	return UserResponse{
		ID:     user.ID,
		Name:   user.Name,
		Email:  user.Email,
		Role:   user.Role,
		Status: user.Status,
	}
}

func ToUsersResponse(user []domain.User) []UserResponse {
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

type UserUpdateRequest struct {
	Name     *string `json:"name,omitempty"`
	Password *string `json:"password,omitempty"`
	Email    *string `json:"email,omitempty"`
	Role     *string `json:"role,omitempty"`
	Status   *string `json:"status,omitempty"`
}

func ToUserUpdateRequest(up UserUpdateRequest) domain.UpdateUser {
	return domain.UpdateUser{
		Name:     up.Name,
		Password: up.Password,
		Email:    up.Email,
		Role:     up.Role,
		Status:   up.Status,
	}
}
