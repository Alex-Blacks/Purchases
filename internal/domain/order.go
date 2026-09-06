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

type OrderListFilter struct {
	GroupIDs    []int
	UserID      *int
	StoreID     *int
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	UpdatedFrom *time.Time
	UpdatedTo   *time.Time
	Limit       int
	Offset      int
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

type OrderItemFindDetails struct {
	StoreID  int
	Store    string
	Quantity int
}

type OrderWithItemDetails struct {
	Order OrderDetails
	Items []OrderItemDetails
}

type OrderItemListFilter struct {
	GroupIDs    []int
	OrderID     *int
	ProductID   *int
	UnitID      *int
	QuantityMin *int
	QuantityMax *int
	Limit       int
	Offset      int
}

func (o OrderWithItemDetails) GetGroupID() int { return o.Order.GroupID }
func (o OrderWithItemDetails) GetID() int      { return o.Order.ID }

type OrderRepository interface {
	Create(ctx context.Context, q Querier, userID, storeID, groupID int) (OrderCreateDetails, error)
	GetByID(ctx context.Context, q Querier, orderID int) (OrderWithItemDetails, error)
	DeleteByID(ctx context.Context, q Querier, orderID int) error
	List(ctx context.Context, q Querier, filter OrderListFilter) ([]OrderDetails, error)
	Count(ctx context.Context, q Querier, filter OrderListFilter) (int, error)
}

type OrderItemRepository interface {
	AddItem(ctx context.Context, q Querier, orderID, productID, UnitID, quantity, groupID int) (OrderItemDetails, error)
	GetItemByOrderAndProduct(ctx context.Context, q Querier, orderID, productID int) (OrderItemDetails, error)
	UpdateItem(ctx context.Context, q Querier, orderID, productID int, updateItem OrderItemUpdate) (OrderItemDetails, error)
	DeleteItemByOrderAndProduct(ctx context.Context, q Querier, orderID, productID int) error
	DeleteAllItems(ctx context.Context, q Querier, orderID int) error
	FindProductInOrders(ctx context.Context, q Querier, productID int, groupID int) ([]OrderItemFindDetails, error)
	List(ctx context.Context, q Querier, filter OrderItemListFilter) ([]OrderItemDetails, error)
	Count(ctx context.Context, q Querier, filter OrderItemListFilter) (int, error)
}
