package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
)

type ServiceAdminProduct struct {
	storage domain.Storage
	product domain.ProductRepository
}

func NewServiceAdminProduct(st domain.Storage, product domain.ProductRepository) *ServiceAdminProduct {
	return &ServiceAdminProduct{storage: st, product: product}
}

// WithTx выполняет функцию в транзакции.
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

// CreateProduct создаёт новый продукт. Доступно только администраторам.
func (s *ServiceAdminProduct) CreateProduct(ctx context.Context, actor policy.Actor, title string, groupID int) (domain.ProductDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.ProductDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("title", title, "group_id", groupID)
	logger.InfoContext(ctx, "creating new product")

	var product domain.ProductDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		// 2. Создание продукта в БД
		product, err = s.product.CreateProduct(ctx, q, title, groupID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to create product", "error", err)
			return fmt.Errorf("create product: %w", err)
		}
		return nil
	}); err != nil {
		return domain.ProductDetails{}, fmt.Errorf("create product: %w", err)
	}

	logger.InfoContext(ctx, "product created successfully", "product_id", product.ID)
	return product, nil
}

// GetProductByID возвращает продукт по ID. Доступно только администраторам.
func (s *ServiceAdminProduct) GetProductByID(ctx context.Context, actor policy.Actor, productID int) (domain.ProductDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.ProductDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("product_id", productID)
	logger.InfoContext(ctx, "getting product by id")

	// 2. Получение продукта из БД (без транзакции)
	product, err := s.product.GetProductByID(ctx, s.storage, productID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get product", "error", err)
		return domain.ProductDetails{}, fmt.Errorf("get product: %w", err)
	}

	logger.InfoContext(ctx, "product retrieved successfully")
	return product, nil
}

// UpdateProductByID обновляет продукт. Доступно только администраторам.
func (s *ServiceAdminProduct) UpdateProductByID(ctx context.Context, actor policy.Actor, productID int, updateProduct domain.ProductUpdate) (domain.ProductDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.ProductDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("product_id", productID)
	logger.InfoContext(ctx, "updating product")

	var product domain.ProductDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		// 2. Обновление продукта в БД
		product, err = s.product.UpdateProductByID(ctx, q, productID, updateProduct)
		if err != nil {
			logger.ErrorContext(ctx, "failed to update product", "error", err)
			return fmt.Errorf("update product: %w", err)
		}
		return nil
	}); err != nil {
		return domain.ProductDetails{}, fmt.Errorf("update product: %w", err)
	}

	logger.InfoContext(ctx, "product updated successfully")
	return product, nil
}

// DeleteProductByID удаляет продукт по ID. Доступно только администраторам.
func (s *ServiceAdminProduct) DeleteProductByID(ctx context.Context, actor policy.Actor, productID int) error {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("product_id", productID)
	logger.InfoContext(ctx, "deleting product")

	// 2. Удаление продукта в транзакции
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		if err := s.product.DeleteProductByID(ctx, q, productID); err != nil {
			logger.ErrorContext(ctx, "failed to delete product", "error", err)
			return fmt.Errorf("delete product: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("delete product: %w", err)
	}

	logger.InfoContext(ctx, "product deleted successfully")
	return nil
}

// ListProducts возвращает список всех продуктов. Доступно только администраторам.
func (s *ServiceAdminProduct) ListProducts(ctx context.Context, actor policy.Actor) ([]domain.ProductDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return []domain.ProductDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx)
	logger.InfoContext(ctx, "listing products")

	// 2. Получение списка продуктов из БД (без транзакции)
	products, err := s.product.ListAdminProducts(ctx, s.storage)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list products", "error", err)
		return []domain.ProductDetails{}, fmt.Errorf("list products: %w", err)
	}

	logger.InfoContext(ctx, "products listed successfully", "count", len(products))
	return products, nil
}

// ---------------------------------------------------------------------------------
// --------------------------------------ALIAS--------------------------------------
// ---------------------------------------------------------------------------------

// CreateProductAlias создаёт алиас для продукта. Доступно только администраторам.
func (s *ServiceAdminProduct) CreateProductAlias(ctx context.Context, actor policy.Actor, productID int, alias string, groupID int) (domain.ProductAliasDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.ProductAliasDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("product_id", productID, "alias", alias)
	logger.InfoContext(ctx, "creating product alias")

	var productAlias domain.ProductAliasDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		// 2. Создание алиаса в БД
		productAlias, err = s.product.CreateProductAlias(ctx, q, productID, alias, groupID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to create product alias", "error", err)
			return fmt.Errorf("create product alias: %w", err)
		}
		return nil
	}); err != nil {
		return domain.ProductAliasDetails{}, fmt.Errorf("create product alias: %w", err)
	}

	logger.InfoContext(ctx, "product alias created successfully", "alias_id", productAlias.ID)
	return productAlias, nil
}

// GetProductAliasByID возвращает алиас по ID. Доступно только администраторам.
func (s *ServiceAdminProduct) GetProductAliasByID(ctx context.Context, actor policy.Actor, aliasID int) (domain.ProductAliasDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.ProductAliasDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("alias_id", aliasID)
	logger.InfoContext(ctx, "getting product alias by id")

	// 2. Получение алиаса из БД (без транзакции)
	alias, err := s.product.GetProductAliasByID(ctx, s.storage, aliasID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get product alias", "error", err)
		return domain.ProductAliasDetails{}, fmt.Errorf("get product alias: %w", err)
	}

	logger.InfoContext(ctx, "product alias retrieved successfully")
	return alias, nil
}

// UpdateProductAliasByID обновляет алиас. Доступно только администраторам.
func (s *ServiceAdminProduct) UpdateProductAliasByID(ctx context.Context, actor policy.Actor, aliasID int, updateAlias domain.ProductAliasUpdate) (domain.ProductAliasDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.ProductAliasDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("alias_id", aliasID)
	logger.InfoContext(ctx, "updating product alias")

	var alias domain.ProductAliasDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		// 2. Обновление алиаса в БД
		alias, err = s.product.UpdateProductAliasByID(ctx, q, aliasID, updateAlias)
		if err != nil {
			logger.ErrorContext(ctx, "failed to update product alias", "error", err)
			return fmt.Errorf("update product alias: %w", err)
		}
		return nil
	}); err != nil {
		return domain.ProductAliasDetails{}, fmt.Errorf("update product alias: %w", err)
	}

	logger.InfoContext(ctx, "product alias updated successfully")
	return alias, nil
}

// DeleteProductAlias удаляет алиас по ID. Доступно только администраторам.
func (s *ServiceAdminProduct) DeleteProductAlias(ctx context.Context, actor policy.Actor, aliasID int) error {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("alias_id", aliasID)
	logger.InfoContext(ctx, "deleting product alias")

	// 2. Удаление алиаса в транзакции
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		if err := s.product.DeleteProductAliasByID(ctx, q, aliasID); err != nil {
			logger.ErrorContext(ctx, "failed to delete product alias", "error", err)
			return fmt.Errorf("delete product alias: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("delete product alias: %w", err)
	}

	logger.InfoContext(ctx, "product alias deleted successfully")
	return nil
}

// ListProductAliases возвращает список алиасов для продукта. Доступно только администраторам.
func (s *ServiceAdminProduct) ListProductAliases(ctx context.Context, actor policy.Actor, productID int) ([]domain.ProductAliasDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return []domain.ProductAliasDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("product_id", productID)
	logger.InfoContext(ctx, "listing product aliases")

	// 2. Получение списка алиасов из БД (без транзакции)
	aliases, err := s.product.ListAdminProductAliases(ctx, s.storage, productID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list product aliases", "error", err)
		return []domain.ProductAliasDetails{}, fmt.Errorf("list product aliases: %w", err)
	}

	logger.InfoContext(ctx, "product aliases listed successfully", "count", len(aliases))
	return aliases, nil
}

// DeleteAllProductAliases удаляет все алиасы продукта. Доступно только администраторам.
func (s *ServiceAdminProduct) DeleteAllProductAliases(ctx context.Context, actor policy.Actor, productID int) error {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("product_id", productID)
	logger.InfoContext(ctx, "deleting all product aliases")

	// 2. Удаление всех алиасов в транзакции
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		if err := s.product.DeleteAllProductAliases(ctx, q, productID); err != nil {
			logger.ErrorContext(ctx, "failed to delete all product aliases", "error", err)
			return fmt.Errorf("delete all product aliases: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("delete all product aliases: %w", err)
	}

	logger.InfoContext(ctx, "all product aliases deleted successfully")
	return nil
}

// FindProductByAlias ищет название продукта по алиасу. Доступно только администраторам.
func (s *ServiceAdminProduct) FindProductByAlias(ctx context.Context, actor policy.Actor, alias string) (string, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return "", policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("alias", alias)
	logger.InfoContext(ctx, "finding product by alias")

	// 2. Поиск продукта по алиасу (без транзакции)
	title, err := s.product.FindAdminProductByAlias(ctx, s.storage, alias)
	if err != nil {
		logger.ErrorContext(ctx, "failed to find product by alias", "error", err)
		return "", fmt.Errorf("find product by alias: %w", err)
	}

	logger.InfoContext(ctx, "product found by alias", "title", title)
	return title, nil
}
