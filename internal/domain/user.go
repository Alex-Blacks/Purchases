package domain

import (
	"context"
)

const (
	UserStatusActive  string = "active"
	UserStatusBlocked string = "blocked"
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

func (u UserDetails) GetGroupID() int { return u.GroupID }
func (u UserDetails) GetID() int      { return u.ID }

type UserRepository interface {
	Create(ctx context.Context, q Querier, name, password_hash, email string, group_id int, role, status string) (UserDetails, error)
	GetByID(ctx context.Context, q Querier, userID int) (UserDetails, error)
	GetByEmail(ctx context.Context, q Querier, email string) (UserDetails, error)
	UpdateByID(ctx context.Context, q Querier, userID int, updateUser UserUpdate) (UserDetails, error)
	DeleteByID(ctx context.Context, q Querier, userID int) error
	ListInGroup(ctx context.Context, q Querier, groupID int) ([]UserDetails, error)
	ListAll(ctx context.Context, q Querier) ([]UserDetails, error)
}
