package domain

import (
	"context"
	"time"
)

type UserStatus string
type UserRole string

const (
	UserStatusActive  UserStatus = "active"
	UserStatusBlocked UserStatus = "blocked"

	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type UserDetails struct {
	ID           int
	Name         string
	PasswordHash string
	Email        string
	GroupID      int
	Group        string
	Role         UserRole
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
type UserUpdate struct {
	Name     *string
	Password *string
	Email    *string
	GroupID  *int
	Role     *UserRole
	Status   *UserStatus
}

type UserListFilter struct {
	GroupIDs    []int
	Role        *UserRole
	Status      *UserStatus
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	UpdatedFrom *time.Time
	UpdatedTo   *time.Time
	Limit       int
	Offset      int
}

func (u UserDetails) GetGroupID() int { return u.GroupID }
func (u UserDetails) GetID() int      { return u.ID }

type UserRepository interface {
	Create(ctx context.Context, q Querier, name, password_hash, email string, group_id int, role UserRole, status UserStatus) (UserDetails, error)
	GetByID(ctx context.Context, q Querier, userID int) (UserDetails, error)
	GetByEmail(ctx context.Context, q Querier, email string) (UserDetails, error)
	UpdateByID(ctx context.Context, q Querier, userID int, updateUser UserUpdate) (UserDetails, error)
	DeleteByID(ctx context.Context, q Querier, userID int) error
	List(ctx context.Context, q Querier, filter UserListFilter) ([]UserDetails, error)
	Count(ctx context.Context, q Querier, filter UserListFilter) (int, error)
}
