package dto

import (
	"time"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

//-------------------------------------------------------------------------------------------------

type StoreRequest struct {
	Name string `json:"name"`
}

type StoreResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

//-------------------------------------------------------------------------------------------------

type CategoryRequest struct {
	Name string `json:"name"`
}

type CategoryResponse struct {
	CategoryID int `json:"categoryId"`
}

//-------------------------------------------------------------------------------------------------

type OrderRequest struct {
	StoreID int `json:"storeId"`
}

type OrderCreateResponse struct {
	ID int `json:"id"`
}

type OrderResponse struct {
	ID         int       `json:"id"`
	User       string    `json:"user"`
	Store      string    `json:"store"`
	ItemsCount int       `json:"itemsCount"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type OrderDetailsResponse struct {
	ID         int                   `json:"id"`
	User       string                `json:"user"`
	Store      string                `json:"store"`
	ItemsCount int                   `json:"itemsCount"`
	CreatedAt  time.Time             `json:"createdAt"`
	UpdatedAt  time.Time             `json:"updatedAt"`
	Items      []ItemDetailsResponse `json:"items"`
}

func ToResponseOrder(o domain.OrderWithItems) OrderDetailsResponse {
	items := make([]ItemDetailsResponse, len(o.Items))
	for i, it := range o.Items {
		items[i] = ItemDetailsResponse{
			ProductID: it.ProductID,
			Title:     it.Title,
			Quantity:  it.Quantity,
		}
	}

	return OrderDetailsResponse{
		ID:         o.Order.ID,
		User:       o.Order.User,
		Store:      o.Order.Store,
		ItemsCount: o.Order.ItemsCount,
		CreatedAt:  o.Order.CreatedAt,
		UpdatedAt:  o.Order.UpdatedAt,
		Items:      items,
	}
}

func ToOrderListResponse(order []domain.Order) []OrderResponse {
	resp := make([]OrderResponse, len(order))

	for i, o := range order {
		resp[i] = OrderResponse{
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

//-------------------------------------------------------------------------------------------------

type ItemRequest struct {
	ProductID int `json:"productId"`
	Quantity  int `json:"quantity"`
}

type ItemUpdateRequest struct {
	Quantity int `json:"quantity"`
}

type ItemDetailsResponse struct {
	ID        int    `json:"id"`
	ProductID int    `json:"productId"`
	Title     string `json:"title"`
	Quantity  int    `json:"quantity"`
}
