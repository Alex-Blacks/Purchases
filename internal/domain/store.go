package domain

import "context"

type Store struct {
	ID   int
	Name string
}

type StoreRepository interface {
	CreateStore(ctx context.Context, q Querier, name string) (Store, error)
	GetStore(ctx context.Context, q Querier, id int) (Store, error)
	DeleteStore(ctx context.Context, q Querier, id int) error
	ListStores(ctx context.Context, q Querier) ([]Store, error)
}
