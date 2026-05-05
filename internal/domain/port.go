package domain

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Storage interface {
	Querier
	BeginTx(ctx context.Context) (Tx, error)
}

type Tx interface {
	Querier
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

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

type OrderDTO struct {
	Id         int
	User       string
	Store      string
	ItemsCount int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
type OrderItemsDTO struct {
	Id       int
	Title    string
	Quantity int
}

type OrderWithItemsDTO struct {
	Order OrderDTO
	Items []OrderItemsDTO
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, q Querier, userID, storeID int) (int, error)
	GetOrder(ctx context.Context, q Querier, userID, orderID int) (OrderWithItemsDTO, error)
	DeleteOrder(ctx context.Context, q Querier, userID, orderID int) error
	ListOrders(ctx context.Context, q Querier, userID int) ([]OrderDTO, error)
}

type OrderItemRepository interface {
	AddItem(ctx context.Context, q Querier, orderID, productID int, qty int) error
	UpdateItem(ctx context.Context, q Querier, orderID, productID int, qty int) error
	DeleteItem(ctx context.Context, q Querier, orderID, productID int) error
	ClearOrder(ctx context.Context, q Querier, orderID int) error
}

type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
