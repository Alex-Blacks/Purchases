package domain

import "context"

type GroupedEntity interface {
	GetGroupID() int
	GetID() int
}

// GenericRepository описывает методы, общие для всех репозиториев
type GenericRepository[T GroupedEntity] interface {
	Create(ctx context.Context, q Querier, params any, groupID int) (T, error)
	GetByID(ctx context.Context, q Querier, id int) (T, error)
	UpdateByID(ctx context.Context, q Querier, id int, updates any) (T, error)
	DeleteByID(ctx context.Context, q Querier, id int) error
	List(ctx context.Context, q Querier, groupIDs []int) ([]T, error)
	ListAll(ctx context.Context, q Querier) ([]T, error)
}
