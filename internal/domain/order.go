package domain

import (
	"context"
	"time"
)

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

type OrderItemCreate struct {
	ProductID int
	UnitID    int
	Quantity  int
}

type OrderItemUpdate struct {
	UnitID   *int
	Quantity *int
}

type OrderWithItemDetails struct {
	Order OrderDetails
	Items []OrderItemDetails
}

func (o OrderWithItemDetails) GetGroupID() int { return o.Order.GroupID }
func (o OrderWithItemDetails) GetID() int      { return o.Order.ID }

type OrderRepository interface {
	Create(ctx context.Context, q Querier, userID, storeID, groupID int) (OrderCreateDetails, error)
	GetByID(ctx context.Context, q Querier, orderID int) (OrderWithItemDetails, error)
	DeleteByID(ctx context.Context, q Querier, orderID int) error
	List(ctx context.Context, q Querier, userID int, groupID int) ([]OrderDetails, error)
	ListAll(ctx context.Context, q Querier) ([]OrderDetails, error)
}

type OrderItemRepository interface {
	GetItemByOrderAndProduct(ctx context.Context, q Querier, orderID, productID int) (OrderItemDetails, error)
	AddItem(ctx context.Context, q Querier, orderID, productID, UnitID, quantity, groupID int) (OrderItemDetails, error)
	DeleteItemByOrderAndProduct(ctx context.Context, q Querier, orderID, productID int) error
	DeleteAllItems(ctx context.Context, q Querier, orderID int) error
	UpdateItem(ctx context.Context, q Querier, orderID, productID int, updateItem OrderItemUpdate) (OrderItemDetails, error)
}
