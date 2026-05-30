package service

import (
	"context"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

type ServiceProduct struct {
	storage domain.Storage
	product domain.ProductRepository
}

func NewServiceProduct(st domain.Storage, product domain.ProductRepository) *ServiceProduct {
	return &ServiceProduct{
		storage: st,
		product: product,
	}
}

func (s *ServiceProduct) CreateProduct(ctx context.Context, title, unit string, categoryID int) (int, error) {
	var productID int
	if err := s.storage.WithTx(ctx, func(q domain.Querier) error {
		var err error
		productID, err = s.product.CreateProduct(ctx, q, title, unit, categoryID)
		return err
	}); err != nil {
		return 0, err
	}
	return productID, nil
}

func (s *ServiceProduct) GetProduct(ctx context.Context, id int) (domain.ProductDetails, error) {
	return s.product.GetProduct(ctx, s.storage, id)
}

func (s *ServiceProduct) DeleteProduct(ctx context.Context, id int) error {
	return s.storage.WithTx(ctx, func(q domain.Querier) error {
		return s.product.DeleteProduct(ctx, q, id)
	})
}

func (s *ServiceProduct) ListProducts(ctx context.Context) ([]domain.ProductDetails, error) {
	return s.product.ListProducts(ctx, s.storage)
}
