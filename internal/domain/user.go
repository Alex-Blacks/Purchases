package domain

import (
	"context"
)

type User struct {
	ID           int
	Name         string
	PasswordHash string
	Email        string
	Role         string
	Status       string
}
type UpdateUser struct {
	Name     *string
	Password *string
	Email    *string
	Role     *string
	Status   *string
}

func (u User) OwnerID() int { return u.ID }

type UserRepository interface {
	CreateUser(ctx context.Context, q Querier, name, password_hash, email, role, status string) (User, error)
	GetUserByID(ctx context.Context, q Querier, userID int) (User, error)
	DeleteUser(ctx context.Context, q Querier, userID int) error
	ListUsers(ctx context.Context, q Querier) ([]User, error)
	UpdateUser(ctx context.Context, q Querier, userID int, updateUser UpdateUser) (User, error)
	GetUserByEmail(ctx context.Context, q Querier, email string) (User, error)
}
