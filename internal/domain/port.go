package domain

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Storage interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)
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
	UpdateAt   time.Time
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

type Order interface {
	CreateOrderTx(ctx context.Context, tx pgx.Tx, userID, storeID int) (int, error)
	GetOrderTx(ctx context.Context, userID, orderID int) ([]OrderWithItemsDTO, error)
	DeleteOrder(ctx context.Context, userID, orderID int) error
	ListOrder(ctx context.Context, userID int) ([]OrderDTO, error)

	AddItem(ctx context.Context, userID, orderID, productID int, qty int) error
	UpdateItem(ctx context.Context, userID, orderID, productID int, qty int) error
	DeleteItem(ctx context.Context, userID, orderID, productID int) error

	ClearOrder(ctx context.Context, userID, orderID int) error
}

type OrderCreator interface {
	CreateOrderTx(ctx context.Context, tx pgx.Tx, userID, storeID int) (int, error)
	GetOrderTx(ctx context.Context, userID, orderID int) ([]OrderWithItemsDTO, error)
}
