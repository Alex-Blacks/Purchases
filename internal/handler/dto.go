package handler

import (
	"time"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

type ResponseOrder struct {
	Id         int       `json:"id"`
	User       string    `json:"user"`
	Store      string    `json:"store"`
	ItemsCount int       `json:"items_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Items      []Item    `json:"items"`
}

type Item struct {
	Id       int    `json:"id"`
	Title    string `json:"title"`
	Quantity int    `json:"quantity"`
}

func ToResponse(o domain.OrderWithItemsDTO) ResponseOrder {
	items := make([]Item, len(o.Items))
	for i, it := range o.Items {
		items[i] = Item{
			Id:       it.Id,
			Title:    it.Title,
			Quantity: it.Quantity,
		}
	}

	return ResponseOrder{
		Id:         o.Order.Id,
		User:       o.Order.User,
		Store:      o.Order.Store,
		ItemsCount: o.Order.ItemsCount,
		CreatedAt:  o.Order.CreatedAt,
		UpdatedAt:  o.Order.UpdatedAt,
		Items:      items,
	}
}
