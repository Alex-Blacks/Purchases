package service

import (
	"github.com/Alex-Blacks/Purchases/internal/domain"
)

type Service struct {
	storage domain.Storage
	order   domain.OrderRepository
	item    domain.OrderItemRepository
	store   domain.StoreRepository
}

func NewService(storage domain.Storage, order domain.OrderRepository, item domain.OrderItemRepository, store domain.StoreRepository) *Service {
	return &Service{
		storage: storage,
		order:   order,
		item:    item,
		store:   store,
	}
}
