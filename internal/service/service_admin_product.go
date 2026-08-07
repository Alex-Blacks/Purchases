package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

type ServiceAdminProduct struct {
	storage domain.Storage
	product domain.ProductRepository
}

func NewServiceAdminProduct(st domain.Storage, product domain.ProductRepository) *ServiceAdminProduct {
	return &ServiceAdminProduct{storage: st, product: product}
}

func (s *ServiceAdminProduct) WithTx(ctx context.Context, fn func(q domain.Querier) error) (err error) {
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

func (s *ServiceAdminProduct) CreateProduct(ctx context.Context, title string, groupID int) (domain.ProductDetails, error) {
	var product domain.ProductDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		product, err = s.product.CreateProduct(ctx, q, title, groupID)
		return err
	}); err != nil {
		return domain.ProductDetails{}, err
	}
	return product, nil
}

func (s *ServiceAdminProduct) GetProductByID(ctx context.Context, productID int) (domain.ProductDetails, error) {
	return s.product.GetProductByID(ctx, s.storage, productID)
}

func (s *ServiceAdminProduct) UpdateProductByID(ctx context.Context, productID int, updateProduct domain.ProductUpdate) (domain.ProductDetails, error) {
	var product domain.ProductDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		product, err = s.product.UpdateProductByID(ctx, q, productID, updateProduct)
		return err
	}); err != nil {
		return domain.ProductDetails{}, err
	}
	return product, nil
}

func (s *ServiceAdminProduct) DeleteProductByID(ctx context.Context, productID int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.product.DeleteProductByID(ctx, q, productID)
	})
}

func (s *ServiceAdminProduct) ListProducts(ctx context.Context) ([]domain.ProductDetails, error) {
	return s.product.ListAdminProducts(ctx, s.storage)
}

// ---------------------------------------------------------------------------------
// --------------------------------------ALIAS--------------------------------------
// ---------------------------------------------------------------------------------

func (s *ServiceAdminProduct) CreateProductAlias(ctx context.Context, productID int, alias string, groupID int) (domain.ProductAliasDetails, error) {
	var productAlias domain.ProductAliasDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		productAlias, err = s.product.CreateProductAlias(ctx, q, productID, alias, groupID)
		return err
	}); err != nil {
		return domain.ProductAliasDetails{}, err
	}
	return productAlias, nil
}

func (s *ServiceAdminProduct) GetProductAliasByID(ctx context.Context, aliasID int) (domain.ProductAliasDetails, error) {
	return s.product.GetProductAliasByID(ctx, s.storage, aliasID)
}

func (s *ServiceAdminProduct) UpdateProductAliasByID(ctx context.Context, aliasID int, updateAlias domain.ProductAliasUpdate) (domain.ProductAliasDetails, error) {
	var alias domain.ProductAliasDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		alias, err = s.product.UpdateProductAliasByID(ctx, q, aliasID, updateAlias)
		return err
	}); err != nil {
		return domain.ProductAliasDetails{}, err
	}
	return alias, nil
}

func (s *ServiceAdminProduct) DeleteProductAlias(ctx context.Context, aliasID int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.product.DeleteProductAliasByID(ctx, q, aliasID)
	})
}

func (s *ServiceAdminProduct) ListProductAliases(ctx context.Context, productID int) ([]domain.ProductAliasDetails, error) {
	return s.product.ListAdminProductAliases(ctx, s.storage, productID)
}

func (s *ServiceAdminProduct) DeleteAllProductAliases(ctx context.Context, productID int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.product.DeleteAllProductAliases(ctx, q, productID)
	})
}

func (s *ServiceAdminProduct) FindProductByAlias(ctx context.Context, alias string) (string, error) {
	return s.product.FindAdminProductByAlias(ctx, s.storage, alias)
}
