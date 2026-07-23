package domain

import (
	"context"
	"time"
)

type OrderDetails struct {
	ID         int
	UserID     int
	User       string
	Store      string
	ItemsCount int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (o OrderDetails) OwnerID() int { return o.UserID }

type OrderItem struct {
	ProductID int
	Quantity  int
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
	CreateOrder(ctx context.Context, q Querier, userID, storeID int) (OrderWithItemDetails, error)
	GetOrder(ctx context.Context, q Querier, userID, orderID int) (OrderWithItemDetails, error)
	DeleteOrder(ctx context.Context, q Querier, userID, orderID int) error
	ListOrders(ctx context.Context, q Querier, userID int) ([]OrderDetails, error)
}

type OrderItemRepository interface {
	GetItemByOrderAndProduct(ctx context.Context, q Querier, orderID, productID int) (OrderItemDetails, error)
	AddItem(ctx context.Context, q Querier, orderID, productID int, quantity int) (OrderItemDetails, error)
	UpdateItem(ctx context.Context, q Querier, orderID, productID int, quantity int) (OrderItemDetails, error)
	DeleteItem(ctx context.Context, q Querier, orderID, productID int) error
	UpsertItem(ctx context.Context, q Querier, orderID, productID, quantity int) error
}
