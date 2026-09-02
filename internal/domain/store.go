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

type StoreListFilter struct {
	GroupIDs []int
	Name     *string
	Limit    int
	Offset   int
}

func (s StoreDetails) GetGroupID() int { return s.GroupID }
func (s StoreDetails) GetID() int      { return s.ID }

type StoreRepository interface {
	Create(ctx context.Context, q Querier, params StoreCreate, groupID int) (StoreDetails, error)
	GetByID(ctx context.Context, q Querier, id int) (StoreDetails, error)
	UpdateByID(ctx context.Context, q Querier, id int, updates StoreUpdate) (StoreDetails, error)
	DeleteByID(ctx context.Context, q Querier, id int) error
	List(ctx context.Context, q Querier, filter StoreListFilter) ([]StoreDetails, error)
	Count(ctx context.Context, q Querier, filter StoreListFilter) (int, error)
}
