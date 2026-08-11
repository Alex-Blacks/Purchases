package domain

import "context"

type GroupDetails struct {
	ID          int
	Name        string
	AdminUserID int
	AdminUser   string
}

type GroupUpdate struct {
	Name        *string
	AdminUserID *int
}

type GroupRepository interface {
	CreateGroup(ctx context.Context, q Querier, name string, adminUserID *int) (GroupDetails, error)
	GetGroupByID(ctx context.Context, q Querier, groupID int) (GroupDetails, error)
	CheckGroupAdmin(ctx context.Context, q Querier, groupID int, adminUserID int) bool
	UpdateGroupByID(ctx context.Context, q Querier, groupID int, updateGroup GroupUpdate) (GroupDetails, error)
	UpdateGroupAdmin(ctx context.Context, q Querier, groupID, adminUserID int) error
	DeleteGroupByID(ctx context.Context, q Querier, groupID int) error
	ListGroups(ctx context.Context, q Querier) ([]GroupDetails, error)
}
