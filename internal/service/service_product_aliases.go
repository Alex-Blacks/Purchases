package service

import (
	"context"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

type ServiceProductAlias struct {
	storage domain.Storage
	alias   domain.ProductAliasRepository
}

func NewServiceProductAlias(st domain.Storage, alias domain.ProductAliasRepository) *ServiceProductAlias {
	return &ServiceProductAlias{
		storage: st,
		alias:   alias,
	}
}

func (s *ServiceProductAlias) CreateProductAlias(ctx context.Context, productID int, alias string) (int, error) {
	var id int
	if err := s.storage.WithTx(ctx, func(q domain.Querier) error {
		var err error
		id, err = s.alias.CreateProductAlias(ctx, q, productID, alias)
		return err
	}); err != nil {
		return 0, err
	}
	return id, nil
}
func (s *ServiceProductAlias) GetProductAlias(ctx context.Context, id int) (domain.ProductAliasDetails, error) {
	return s.alias.GetProductAlias(ctx, s.storage, id)
}
func (s *ServiceProductAlias) DeleteProductAlias(ctx context.Context, id int) error {
	return s.storage.WithTx(ctx, func(q domain.Querier) error {
		return s.alias.DeleteProductAlias(ctx, q, id)
	})
}
func (s *ServiceProductAlias) ListProductAliases(ctx context.Context, productID int) ([]domain.ProductAliasDetails, error) {
	return s.alias.ListProductAliases(ctx, s.storage, productID)
}
func (s *ServiceProductAlias) DeleteAllProductAliases(ctx context.Context, productID int) error {
	return s.storage.WithTx(ctx, func(q domain.Querier) error {
		return s.alias.DeleteAllProductAliases(ctx, q, productID)
	})
}
func (s *ServiceProductAlias) FindProductByAlias(ctx context.Context, alias string) (string, error) {
	return s.alias.FindProductByAlias(ctx, s.storage, alias)
}
