package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/policy"
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

func (s *ServiceProduct) AccessReadProduct(ctx context.Context, actor policy.Actor, productID int) (domain.ProductDetails, error) {
	product, err := s.product.GetProductByID(ctx, s.storage, productID)
	if err != nil {
		return domain.ProductDetails{}, err
	}
	if err := policy.CanGroupAccessForReading(actor, product); err != nil {
		return domain.ProductDetails{}, err
	}
	return product, nil
}

func (s *ServiceProduct) AccessWriteProduct(ctx context.Context, actor policy.Actor, productID int) error {
	product, err := s.product.GetProductByID(ctx, s.storage, productID)
	if err != nil {
		return err
	}
	if err := policy.CanGroupAccessForModify(actor, product); err != nil {
		return err
	}
	return nil
}

func (s *ServiceProduct) AccessReadProductAlias(ctx context.Context, actor policy.Actor, aliasID int) (domain.ProductAliasDetails, error) {
	alias, err := s.product.GetProductAliasByID(ctx, s.storage, aliasID)
	if err != nil {
		return domain.ProductAliasDetails{}, err
	}
	if err := policy.CanGroupAccessForReading(actor, alias); err != nil {
		return domain.ProductAliasDetails{}, err
	}
	return alias, nil
}

func (s *ServiceProduct) AccessWriteProductAlias(ctx context.Context, actor policy.Actor, aliasID int) error {
	alias, err := s.product.GetProductAliasByID(ctx, s.storage, aliasID)
	if err != nil {
		return err
	}
	if err := policy.CanGroupAccessForModify(actor, alias); err != nil {
		return err
	}
	return nil
}

func (s *ServiceProduct) CreateProduct(ctx context.Context, actor policy.Actor, title string) (domain.ProductDetails, error) {
	var product domain.ProductDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		product, err = s.product.CreateProduct(ctx, q, title, actor.GroupID)
		return err
	}); err != nil {
		return domain.ProductDetails{}, err
	}
	return product, nil
}

func (s *ServiceProduct) GetProductByID(ctx context.Context, actor policy.Actor, productID int) (domain.ProductDetails, error) {
	product, err := s.AccessReadProduct(ctx, actor, productID)
	if err != nil {
		return domain.ProductDetails{}, err
	}
	return product, nil
}

func (s *ServiceProduct) UpdateProductByID(ctx context.Context, actor policy.Actor, productID int, updateProduct domain.ProductUpdate) (domain.ProductDetails, error) {
	if err := s.AccessWriteProduct(ctx, actor, productID); err != nil {
		return domain.ProductDetails{}, err
	}
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

func (s *ServiceProduct) DeleteProductByID(ctx context.Context, actor policy.Actor, productID int) error {
	if err := s.AccessWriteProduct(ctx, actor, productID); err != nil {
		return err
	}
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.product.DeleteProductByID(ctx, q, productID)
	})
}

func (s *ServiceProduct) ListProducts(ctx context.Context, actor policy.Actor) ([]domain.ProductDetails, error) {
	return s.product.ListProducts(ctx, s.storage, []int{actor.GroupID, policy.CommonGroupID})
}

// ---------------------------------------------------------------------------------
// --------------------------------------ALIAS--------------------------------------
// ---------------------------------------------------------------------------------

func (s *ServiceProduct) CreateProductAlias(ctx context.Context, actor policy.Actor, productID int, alias string) (domain.ProductAliasDetails, error) {
	if err := s.AccessWriteProduct(ctx, actor, productID); err != nil {
		return domain.ProductAliasDetails{}, err
	}
	var productAlias domain.ProductAliasDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		productAlias, err = s.product.CreateProductAlias(ctx, q, productID, alias, actor.GroupID)
		return err
	}); err != nil {
		return domain.ProductAliasDetails{}, err
	}
	return productAlias, nil
}

func (s *ServiceProduct) GetProductAliasByID(ctx context.Context, actor policy.Actor, aliasID int) (domain.ProductAliasDetails, error) {
	alias, err := s.AccessReadProductAlias(ctx, actor, aliasID)
	if err != nil {
		return domain.ProductAliasDetails{}, err
	}
	return alias, nil
}

func (s *ServiceProduct) UpdateProductAliasByID(ctx context.Context, actor policy.Actor, aliasID int, newAlias string) (domain.ProductAliasDetails, error) {
	if err := s.AccessWriteProductAlias(ctx, actor, aliasID); err != nil {
		return domain.ProductAliasDetails{}, err
	}
	var alias domain.ProductAliasDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		alias, err = s.product.UpdateProductAliasByID(ctx, q, aliasID, domain.ProductAliasUpdate{Alias: &newAlias})
		return err
	}); err != nil {
		return domain.ProductAliasDetails{}, err
	}
	return alias, nil
}

func (s *ServiceProduct) DeleteProductAlias(ctx context.Context, actor policy.Actor, aliasID int) error {
	if err := s.AccessWriteProductAlias(ctx, actor, aliasID); err != nil {
		return err
	}
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.product.DeleteProductAliasByID(ctx, q, aliasID)
	})
}

func (s *ServiceProduct) ListProductAliases(ctx context.Context, actor policy.Actor, productID int) ([]domain.ProductAliasDetails, error) {
	if _, err := s.AccessReadProduct(ctx, actor, productID); err != nil {
		return []domain.ProductAliasDetails{}, err
	}
	return s.product.ListProductAliases(ctx, s.storage, productID, []int{actor.GroupID, policy.CommonGroupID})
}

func (s *ServiceProduct) DeleteAllProductAliases(ctx context.Context, actor policy.Actor, productID int) error {
	if err := s.AccessWriteProduct(ctx, actor, productID); err != nil {
		return err
	}
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.product.DeleteAllProductAliases(ctx, q, productID)
	})
}

func (s *ServiceProduct) FindProductByAlias(ctx context.Context, actor policy.Actor, alias string) (string, error) {
	return s.product.FindProductByAlias(ctx, s.storage, alias, []int{actor.GroupID, policy.CommonGroupID})
}
