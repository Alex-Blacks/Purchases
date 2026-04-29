package domain

import "context"

type StoreDTO struct {
	Id   int
	Name string
}

type Store interface {
	CreateStore(ctx context.Context, name string) error
	GetStoreById(ctx context.Context, id int) (string, error)
	DeleteStore(ctx context.Context, id int) error
	ListStore(ctx context.Context) ([]StoreDTO, error)
}
