package service

import (
	"context"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

func (s *Service) CreateProductAlias(ctx context.Context, productID int, alias string) (int, error) {
	var id int
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		id, err = s.productAliases.CreateProductAlias(ctx, q, productID, alias)
		return err
	}); err != nil {
		return 0, err
	}
	return id, nil
}
func (s *Service) GetProductAlias(ctx context.Context, id int) (domain.ProductAliasDetails, error) {
	return s.productAliases.GetProductAlias(ctx, s.storage, id)
}
func (s *Service) DeleteProductAlias(ctx context.Context, id int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.productAliases.DeleteProductAlias(ctx, q, id)
	})
}
func (s *Service) ListProductAliases(ctx context.Context, productID int) ([]domain.ProductAliasDetails, error) {
	return s.productAliases.ListProductAliases(ctx, s.storage, productID)
}
func (s *Service) DeleteAllProductAliases(ctx context.Context, productID int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.productAliases.DeleteAllProductAliases(ctx, q, productID)
	})
}
func (s *Service) FindProductByAlias(ctx context.Context, alias string) (int, error) {
	return s.productAliases.FindProductByAlias(ctx, s.storage, alias)
}
