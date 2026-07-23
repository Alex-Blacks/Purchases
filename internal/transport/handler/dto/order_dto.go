package dto

import (
	"time"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

type OrderRequest struct {
	StoreID int `json:"storeId"`
}

type OrderDetailsResponse struct {
	ID         int       `json:"id"`
	User       string    `json:"user"`
	Store      string    `json:"store"`
	ItemsCount int       `json:"itemsCount"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type OrderWithItemDetailsResponse struct {
	ID         int                   `json:"id"`
	User       string                `json:"user"`
	Store      string                `json:"store"`
	ItemsCount int                   `json:"itemsCount"`
	CreatedAt  time.Time             `json:"createdAt"`
	UpdatedAt  time.Time             `json:"updatedAt"`
	Items      []ItemDetailsResponse `json:"items"`
}

type ItemRequest struct {
	ProductID int `json:"productId"`
	Quantity  int `json:"quantity"`
}

type ListItemsRequest struct {
	Items []ItemRequest `json:"items"`
}

func ToItemsRequest(items ListItemsRequest) []domain.OrderItem {
	resp := make([]domain.OrderItem, len(items.Items))

	for id, i := range items.Items {
		resp[id] = domain.OrderItem{
			ProductID: i.ProductID,
			Quantity:  i.Quantity,
		}
	}

	return resp
}

type ItemUpdateRequest struct {
	Quantity int `json:"quantity"`
}

type ItemDetailsResponse struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Quantity int    `json:"quantity"`
}

func ToResponseOrder(o domain.OrderWithItemDetails) OrderWithItemDetailsResponse {
	items := make([]ItemDetailsResponse, len(o.Items))
	for i, it := range o.Items {
		items[i] = ItemDetailsResponse{
			Title:    it.Title,
			Quantity: it.Quantity,
		}
	}

	return OrderWithItemDetailsResponse{
		ID:         o.Order.ID,
		User:       o.Order.User,
		Store:      o.Order.Store,
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
			ID:         o.ID,
			User:       o.User,
			Store:      o.Store,
			ItemsCount: o.ItemsCount,
			CreatedAt:  o.CreatedAt,
			UpdatedAt:  o.UpdatedAt,
		}
	}
	return resp
}
