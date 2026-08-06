package domain

import "context"

type StoreDetails struct {
	ID      int
	Name    string
	GroupID int
	Group   string
}

type StoreUpdate struct {
	Name *string
}

func (s StoreDetails) GetGroupID() int { return s.GroupID }

type StoreRepository interface {
	CreateStore(ctx context.Context, q Querier, name string, groupID int) (StoreDetails, error)
	GetStoreByID(ctx context.Context, q Querier, storeID int) (StoreDetails, error)
	UpdateStoreByID(ctx context.Context, q Querier, storeID int, updateStore StoreUpdate) (StoreDetails, error)
	DeleteStoreByID(ctx context.Context, q Querier, storeID int) error
	ListStores(ctx context.Context, q Querier, groupID []int) ([]StoreDetails, error)
}
