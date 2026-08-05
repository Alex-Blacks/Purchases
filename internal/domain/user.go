package domain

import (
	"context"
)

type UserDetails struct {
	ID           int
	Name         string
	PasswordHash string
	Email        string
	GroupID      int
	Group        string
	Role         string
	Status       string
}
type UserUpdate struct {
	Name     *string
	Password *string
	Email    *string
	GroupID  *int
	Role     *string
	Status   *string
}

func (u UserDetails) OwnerID() int { return u.ID }

type UserRepository interface {
	CreateUser(ctx context.Context, q Querier, name, password_hash, email string, group_id int, role, status string) (UserDetails, error)
	GetUserByID(ctx context.Context, q Querier, userID int) (UserDetails, error)
	UpdateUserByID(ctx context.Context, q Querier, userID int, updateUser UserUpdate) (UserDetails, error)
	DeleteUserByID(ctx context.Context, q Querier, userID int) error
	ListUsers(ctx context.Context, q Querier) ([]UserDetails, error)
	GetUserByEmail(ctx context.Context, q Querier, email string) (UserDetails, error)
}
