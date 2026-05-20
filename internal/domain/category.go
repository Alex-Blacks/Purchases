package domain

import "context"

type Category struct {
	ID   int
	Name string
}

type CategoryRepository interface {
	CreateCategory(ctx context.Context, q Querier, name string) (int, error)
	GetCategory(ctx context.Context, q Querier, id int) (Category, error)
	DeleteCategory(ctx context.Context, q Querier, id int) error
	ListCategories(ctx context.Context, q Querier) ([]Category, error)
}
