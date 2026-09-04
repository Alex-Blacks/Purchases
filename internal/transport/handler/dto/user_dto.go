package dto

import (
	"time"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

// UserRequest используется для создания пользователя.
type UserRequest struct {
	Name     string           `json:"name" validate:"required,min=1,max=50"`
	Password string           `json:"password" validate:"required,min=8,max=100"`
	Email    string           `json:"email" validate:"required,email"`
	Role     *domain.UserRole `json:"role,omitempty" validate:"oneof=admin user"`
}

// UserUpdateRequest используется для обновления пользователя.
type UserUpdateRequest struct {
	Name     *string            `json:"name,omitempty" validate:"min=1,max=50"`
	Password *string            `json:"password,omitempty" validate:"min=8,max=100"`
	Email    *string            `json:"email,omitempty" validate:"email"`
	Role     *domain.UserRole   `json:"role,omitempty" validate:"oneof=admin user"`
	Status   *domain.UserStatus `json:"status,omitempty" validate:"oneof=active blocked"`
}

// UserResponse возвращает информацию о пользователе.
type UserResponse struct {
	ID     int               `json:"id"`
	Name   string            `json:"name"`
	Email  string            `json:"email"`
	Role   domain.UserRole   `json:"role"`
	Status domain.UserStatus `json:"status"`
}

// UserFilterQuery – структура для биндинга query-параметров
type UserFilterQuery struct {
	GroupIDs    []int      `form:"group_ids" validate:"dive,int,gt=0"`          // ?group_ids=1&group_ids=2
	Role        string     `form:"role" validate:"oneof=admin user"`            // admin / user
	Status      string     `form:"status" validate:"oneof=active blocked"`      // active / blocked
	CreatedFrom *time.Time `form:"created_from" validate:"daterange=CreatedTo"` // RFC3339
	CreatedTo   *time.Time `form:"created_to"`
	UpdatedFrom *time.Time `form:"updated_from" validate:"daterange=UpdatedTo"`
	UpdatedTo   *time.Time `form:"updated_to"`
	Limit       int        `form:"limit" default:"10" validate:"required,min=1,max=100"`
	Offset      int        `form:"offset" default:"0" validate:"min=0"`
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

// ToDomainFilter преобразует dto.UserFilterQuery в domain.UserListFilter.
func (q UserFilterQuery) ToDomainFilter() domain.UserListFilter {
	filter := domain.UserListFilter{
		GroupIDs:    q.GroupIDs,
		CreatedFrom: q.CreatedFrom,
		CreatedTo:   q.CreatedTo,
		UpdatedFrom: q.UpdatedFrom,
		UpdatedTo:   q.UpdatedTo,
		Limit:       q.Limit,
		Offset:      q.Offset,
	}

	if q.Role != "" {
		role := domain.UserRole(q.Role)
		filter.Role = &role
	}
	if q.Status != "" {
		status := domain.UserStatus(q.Status)
		filter.Status = &status
	}
	return filter
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
