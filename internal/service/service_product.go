package service

import (
	"context"
	"fmt"

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

func (s *ServiceProduct) WithTx(ctx context.Context, fn func(q domain.Querier) error) (err error) {
	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = fmt.Errorf("tx err: %v, rollback err: %w", err, rollbackErr)
			}
			return
		}

		if commitErr := tx.Commit(ctx); commitErr != nil {
			err = fmt.Errorf("commit err: %w", commitErr)
		}
	}()

	err = fn(tx)
	return err
}

func (s *ServiceProduct) CreateProduct(ctx context.Context, title, unit string, categoryID int) (domain.ProductDetails, error) {
	var product domain.ProductDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		product, err = s.product.CreateProduct(ctx, q, title, unit, categoryID)
		return err
	}); err != nil {
		return domain.ProductDetails{}, err
	}
	return product, nil
}

func (s *ServiceProduct) GetProduct(ctx context.Context, id int) (domain.ProductDetails, error) {
	return s.product.GetProduct(ctx, s.storage, id)
}

func (s *ServiceProduct) DeleteProduct(ctx context.Context, id int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.product.DeleteProduct(ctx, q, id)
	})
}

func (s *ServiceProduct) ListProducts(ctx context.Context) ([]domain.ProductDetails, error) {
	return s.product.ListProducts(ctx, s.storage)
}

func (s *ServiceProduct) CreateProductAlias(ctx context.Context, productID int, alias string) (domain.ProductAliasDetails, error) {
	var productAlias domain.ProductAliasDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		productAlias, err = s.product.CreateProductAlias(ctx, q, productID, alias)
		return err
	}); err != nil {
		return domain.ProductAliasDetails{}, err
	}
	return productAlias, nil
}
func (s *ServiceProduct) GetProductAlias(ctx context.Context, id int) (domain.ProductAliasDetails, error) {
	return s.product.GetProductAlias(ctx, s.storage, id)
}
func (s *ServiceProduct) DeleteProductAlias(ctx context.Context, id int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.product.DeleteProductAlias(ctx, q, id)
	})
}
func (s *ServiceProduct) ListProductAliases(ctx context.Context, productID int) ([]domain.ProductAliasDetails, error) {
	return s.product.ListProductAliases(ctx, s.storage, productID)
}
func (s *ServiceProduct) DeleteAllProductAliases(ctx context.Context, productID int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.product.DeleteAllProductAliases(ctx, q, productID)
	})
}
func (s *ServiceProduct) FindProductByAlias(ctx context.Context, alias string) (string, error) {
	return s.product.FindProductByAlias(ctx, s.storage, alias)
}
