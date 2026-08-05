package domain

import (
	"context"
	"time"
)

func (o OrderDetails) OwnerID() int { return o.UserID }

type OrderDetails struct {
	ID         int
	UserID     int
	User       string
	StoreID    int
	Store      string
	GroupID    int
	Group      string
	ItemsCount int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type OrderCreateDetails struct {
	ID        int
	UserID    int
	User      string
	StoreID   int
	Store     string
	GroupID   int
	Group     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type OrderItemDetails struct {
	ID        int
	OrderID   int
	ProductID int
	Title     string
	UnitID    int
	Unit      string
	Quantity  int
	GroupID   int
	Group     string
}

type OrderItemUpdate struct {
	UnitID   *int
	Quantity *int
}

type CreateOrderItem struct {
	ProductID int
	UnitID    int
	Quantity  int
}

type OrderWithItemDetails struct {
	Order OrderDetails
	Items []OrderItemDetails
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, q Querier, userID, storeID, groupID int) (OrderCreateDetails, error)
	GetOrderByID(ctx context.Context, q Querier, orderID int) (OrderWithItemDetails, error)
	DeleteOrderByID(ctx context.Context, q Querier, orderID int) error
	ListOrders(ctx context.Context, q Querier, userID int, groupID int) ([]OrderDetails, error)
}

type OrderItemRepository interface {
	GetItemByOrderAndProduct(ctx context.Context, q Querier, orderID, productID int) (OrderItemDetails, error)
	AddItem(ctx context.Context, q Querier, orderID, productID, UnitID, quantity, groupID int) (OrderItemDetails, error)
	DeleteItemByID(ctx context.Context, q Querier, orderID, productID int) error
	DeleteAllItems(ctx context.Context, q Querier, orderID int) error
	UpdateItem(ctx context.Context, q Querier, orderID, productID int, updateItem OrderItemUpdate) (OrderItemDetails, error)
}
