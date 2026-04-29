package service

import (
	"github.com/Alex-Blacks/Purchases/internal/domain"
)

type Service struct {
	storage domain.Storage
	store   domain.Store
	order   domain.Order
}

func NewService(store domain.Store) *Service {
	return &Service{store: store}
}
