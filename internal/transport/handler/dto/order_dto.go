package dto

import (
	"time"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

// OrderRequest используется для создания заказа.
type OrderRequest struct {
	StoreID int  `json:"storeId" validate:"required,gt=0"`
	GroupID *int `json:"groupId,omitempty" validate:"gt=0"` // опционально, для админов
}

// OrderDetailsResponse возвращает детальную информацию о заказе.
type OrderDetailsResponse struct {
	ID        int       `json:"id"`
	UserID    int       `json:"userId"`
	User      string    `json:"user"`
	StoreID   int       `json:"storeId"`
	Store     string    `json:"store"`
	GroupID   int       `json:"groupId"`
	Group     string    `json:"group"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ToOrderResponse преобразует domain.OrderCreateDetails в OrderDetailsResponse.
func ToOrderResponse(order domain.OrderCreateDetails) OrderDetailsResponse {
	return OrderDetailsResponse{
		ID:        order.ID,
		UserID:    order.UserID,
		User:      order.User,
		StoreID:   order.StoreID,
		Store:     order.Store,
		GroupID:   order.GroupID,
		Group:     order.Group,
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
	}
}

// ItemRequest используется для добавления элемента.
type ItemRequest struct {
	ProductID int  `json:"productId" validate:"required,gt=0"`
	UnitID    int  `json:"unitId" validate:"required,gt=0"`
	Quantity  int  `json:"quantity" validate:"required,gt=0"`
	GroupID   *int `json:"groupId,omitempty" validate:"gt=0"` // опционально, для админов
}

// ItemUpdateRequest используется для обновления элемента в заказе.
type ItemUpdateRequest struct {
	UnitID   *int `json:"unitId,omitempty" validate:"gt=0"`
	Quantity *int `json:"quantity,omitempty" validate:"gt=0"`
}

// ListItemsRequest используется для добавления списка элементов.
type ListItemsRequest struct {
	Items   []ItemRequest `json:"items" validate:"required, dive"`
	GroupID *int          `json:"groupId,omitempty" validate:"gt=0"` // опционально, для админов
}

type ItemDetailsResponse struct {
	ID        int    `json:"id"`
	OrderID   int    `json:"orderId"`
	ProductID int    `json:"productId"`
	Title     string `json:"title"`
	UnitID    int    `json:"unitId"`
	Unit      string `json:"unit"`
	Quantity  int    `json:"quantity"`
}

func ToItemResponse(item domain.OrderItemDetails) ItemDetailsResponse {
	return ItemDetailsResponse{
		ID:        item.ID,
		OrderID:   item.OrderID,
		ProductID: item.ProductID,
		Title:     item.Title,
		UnitID:    item.UnitID,
		Unit:      item.Unit,
		Quantity:  item.Quantity,
	}
}

func ToItemListRequest(items ListItemsRequest) []domain.OrderItemCreate {
	resp := make([]domain.OrderItemCreate, len(items.Items))

	for id, i := range items.Items {
		resp[id] = domain.OrderItemCreate{
			ProductID: i.ProductID,
			UnitID:    i.UnitID,
			Quantity:  i.Quantity,
		}
	}

	return resp
}

func ToItemUpdateRequest(item ItemUpdateRequest) domain.OrderItemUpdate {
	return domain.OrderItemUpdate{
		UnitID:   item.UnitID,
		Quantity: item.Quantity,
	}
}

type OrderWithItemDetailsResponse struct {
	ID         int                   `json:"id"`
	UserID     int                   `json:"userId"`
	User       string                `json:"user"`
	StoreID    int                   `json:"storeId"`
	Store      string                `json:"store"`
	GroupID    int                   `json:"groupId"`
	Group      string                `json:"group"`
	ItemsCount int                   `json:"itemsCount"`
	CreatedAt  time.Time             `json:"createdAt"`
	UpdatedAt  time.Time             `json:"updatedAt"`
	Items      []ItemDetailsResponse `json:"items"`
}

func ToOrderWithItemResponse(o domain.OrderWithItemDetails) OrderWithItemDetailsResponse {
	items := make([]ItemDetailsResponse, len(o.Items))
	for i, it := range o.Items {
		items[i] = ItemDetailsResponse{
			ID:        it.ID,
			OrderID:   it.OrderID,
			ProductID: it.ProductID,
			Title:     it.Title,
			UnitID:    it.UnitID,
			Unit:      it.Unit,
			Quantity:  it.Quantity,
		}
	}

	return OrderWithItemDetailsResponse{
		ID:         o.Order.ID,
		UserID:     o.Order.UserID,
		User:       o.Order.User,
		StoreID:    o.Order.StoreID,
		Store:      o.Order.Store,
		GroupID:    o.Order.GroupID,
		Group:      o.Order.Group,
		ItemsCount: o.Order.ItemsCount,
		CreatedAt:  o.Order.CreatedAt,
		UpdatedAt:  o.Order.UpdatedAt,
		Items:      items,
	}
}

func ToOrderListResponse(order []domain.OrderDetails) []OrderDetailsResponse {
	resp := make([]OrderDetailsResponse, len(order))

	for i, o := range order {
		resp[i] = OrderDetailsResponse{
			ID:        o.ID,
			UserID:    o.UserID,
			User:      o.User,
			StoreID:   o.StoreID,
			Store:     o.Store,
			GroupID:   o.GroupID,
			Group:     o.Group,
			CreatedAt: o.CreatedAt,
			UpdatedAt: o.UpdatedAt,
		}
	}
	return resp
}

type OrderItemFindResponse struct {
	StoreID  int    `json:"storeId"`
	Store    string `json:"store"`
	Quantity int    `json:"quantity"`
}

func ToOrderItemFindResponse(stores []domain.OrderItemFindDetails) []OrderItemFindResponse {
	resp := make([]OrderItemFindResponse, len(stores))

	for i, s := range stores {
		resp[i] = OrderItemFindResponse{
			StoreID:  s.StoreID,
			Store:    s.Store,
			Quantity: s.Quantity,
		}
	}

	return resp
}
