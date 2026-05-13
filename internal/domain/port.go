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

// ---------------------------------------------------------------------------------------------------------------------------------

type Store struct {
	ID   int
	Name string
}

type StoreRepository interface {
	CreateStore(ctx context.Context, q Querier, name string) (int, error)
	GetStore(ctx context.Context, q Querier, id int) (Store, error)
	DeleteStore(ctx context.Context, q Querier, id int) error
	ListStores(ctx context.Context, q Querier) ([]Store, error)
}

// ---------------------------------------------------------------------------------------------------------------------------------

type Category struct {
	ID   int
	Name string
}

type CategoryRepositoriy interface {
	CreateCategory(ctx context.Context, q Querier, name string) (int, error)
	GetCategory(ctx context.Context, q Querier, id int) (Category, error)
	DeleteCategory(ctx context.Context, q Querier, id int) error
	ListCategories(ctx context.Context, q Querier) ([]Category, error)
}

// ---------------------------------------------------------------------------------------------------------------------------------

type OrderDetails struct {
	ID         int
	User       string
	Store      string
	ItemsCount int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
type OrderItemDetails struct {
	ID        int
	ProductID int
	Title     string
	Quantity  int
}

type OrderWithItemDetails struct {
	Order OrderDetails
	Items []OrderItemDetails
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, q Querier, userID, storeID int) (int, error)
	GetOrder(ctx context.Context, q Querier, userID, orderID int) (OrderWithItemDetails, error)
	DeleteOrder(ctx context.Context, q Querier, userID, orderID int) error
	ListOrders(ctx context.Context, q Querier, userID int) ([]OrderDetails, error)
}

type OrderItemRepository interface {
	AddItem(ctx context.Context, q Querier, orderID, productID int, qty int) (OrderItemDetails, error)
	UpdateItem(ctx context.Context, q Querier, orderID, productID int, qty int) (OrderItemDetails, error)
	DeleteItem(ctx context.Context, q Querier, orderID, productID int) error
}

type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

//---------------------------------------------------------------------------------------------------------------------------------
