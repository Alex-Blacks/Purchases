package dto

import (
	"time"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

type StoreResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type OrderResponse struct {
	ID         int            `json:"id"`
	User       string         `json:"user"`
	Store      string         `json:"store"`
	ItemsCount int            `json:"items_count"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	Items      []ItemResponse `json:"items"`
}

type ItemResponse struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Quantity int    `json:"quantity"`
}

func ToResponseOrder(o domain.OrderWithItems) OrderResponse {
	items := make([]ItemResponse, len(o.Items))
	for i, it := range o.Items {
		items[i] = ItemResponse{
			ID:       it.ID,
			Title:    it.Title,
			Quantity: it.Quantity,
		}
	}

	return OrderResponse{
		ID:         o.Order.ID,
		User:       o.Order.User,
		Store:      o.Order.Store,
		ItemsCount: o.Order.ItemsCount,
		CreatedAt:  o.Order.CreatedAt,
		UpdatedAt:  o.Order.UpdatedAt,
		Items:      items,
	}
}
