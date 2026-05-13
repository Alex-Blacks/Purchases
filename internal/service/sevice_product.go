package service

import (
	"context"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

func (s *Service) CreateProduct(ctx context.Context, title, unit string, categoryID int) (int, error) {
	var productID int
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		productID, err = s.product.CreateProduct(ctx, q, title, unit, categoryID)
		return err
	}); err != nil {
		return 0, err
	}
	return productID, nil
}

func (s *Service) GetProduct(ctx context.Context, id int) (domain.ProductDetails, error) {
	return s.product.GetProduct(ctx, s.storage, id)
}

func (s *Service) DeleteProduct(ctx context.Context, id int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.product.DeleteProduct(ctx, q, id)
	})
}

func (s *Service) ListProducts(ctx context.Context) ([]domain.ProductDetails, error) {
	return s.product.ListProducts(ctx, s.storage)
}
