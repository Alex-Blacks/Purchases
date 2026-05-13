package service

import (
	"github.com/Alex-Blacks/Purchases/internal/domain"
)

type Service struct {
	storage  domain.Storage
	order    domain.OrderRepository
	item     domain.OrderItemRepository
	store    domain.StoreRepository
	category domain.CategoryRepositoriy
	product  domain.ProductRepository
}

func NewService(
	storage domain.Storage,
	order domain.OrderRepository,
	item domain.OrderItemRepository,
	store domain.StoreRepository,
	category domain.CategoryRepositoriy,
	product domain.ProductRepository,
) *Service {
	return &Service{
		storage:  storage,
		order:    order,
		item:     item,
		store:    store,
		category: category,
		product:  product,
	}
}
