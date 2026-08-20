package domain

import "context"

type StoreCreate struct {
	Name string
}

type StoreUpdate struct {
	Name *string
}

type StoreDetails struct {
	ID      int
	Name    string
	GroupID int
	Group   string
}

func (s StoreDetails) GetGroupID() int { return s.GroupID }
func (s StoreDetails) GetID() int      { return s.ID }

type StoreRepository interface {
	Create(ctx context.Context, q Querier, params any, groupID int) (StoreDetails, error)
	GetByID(ctx context.Context, q Querier, id int) (StoreDetails, error)
	UpdateByID(ctx context.Context, q Querier, id int, updates any) (StoreDetails, error)
	DeleteByID(ctx context.Context, q Querier, id int) error
	List(ctx context.Context, q Querier, groupIDs []int) ([]StoreDetails, error)
	ListAll(ctx context.Context, q Querier) ([]StoreDetails, error)
}
