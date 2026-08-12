package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
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

// WithTx выполняет функцию в транзакции.
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

// AccessReadProduct проверяет доступ к чтению продукта и возвращает его.
func (s *ServiceProduct) AccessReadProduct(ctx context.Context, actor policy.Actor, productID int) (domain.ProductDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("product_id", productID)
	logger.InfoContext(ctx, "checking read access for product")

	// 1. Получение продукта из БД
	product, err := s.product.GetProductByID(ctx, s.storage, productID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get product", "error", err)
		return domain.ProductDetails{}, fmt.Errorf("get product: %w", err)
	}

	// 2. Проверка прав на чтение через policy
	if err := policy.CanGroupAccessForReading(actor, product); err != nil {
		logger.WarnContext(ctx, "read access denied", "error", err)
		return domain.ProductDetails{}, err
	}

	logger.InfoContext(ctx, "read access granted")
	return product, nil
}

// AccessWriteProduct проверяет доступ к изменению продукта.
func (s *ServiceProduct) AccessWriteProduct(ctx context.Context, actor policy.Actor, productID int) error {
	logger := logging.LoggerFromContext(ctx).With("product_id", productID)
	logger.InfoContext(ctx, "checking write access for product")

	// 1. Получение продукта из БД
	product, err := s.product.GetProductByID(ctx, s.storage, productID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get product", "error", err)
		return fmt.Errorf("get product: %w", err)
	}

	// 2. Проверка прав на изменение через policy
	if err := policy.CanGroupAccessForModify(actor, product); err != nil {
		logger.WarnContext(ctx, "write access denied", "error", err)
		return err
	}

	logger.InfoContext(ctx, "write access granted")
	return nil
}

// AccessReadProductAlias проверяет доступ к чтению алиаса продукта и возвращает его.
func (s *ServiceProduct) AccessReadProductAlias(ctx context.Context, actor policy.Actor, aliasID int) (domain.ProductAliasDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("alias_id", aliasID)
	logger.InfoContext(ctx, "checking read access for product alias")

	// 1. Получение алиаса из БД
	alias, err := s.product.GetProductAliasByID(ctx, s.storage, aliasID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get product alias", "error", err)
		return domain.ProductAliasDetails{}, fmt.Errorf("get product alias: %w", err)
	}

	// 2. Проверка прав на чтение через policy
	if err := policy.CanGroupAccessForReading(actor, alias); err != nil {
		logger.WarnContext(ctx, "read access denied", "error", err)
		return domain.ProductAliasDetails{}, err
	}

	logger.InfoContext(ctx, "read access granted")
	return alias, nil
}

// AccessWriteProductAlias проверяет доступ к изменению алиаса продукта.
func (s *ServiceProduct) AccessWriteProductAlias(ctx context.Context, actor policy.Actor, aliasID int) error {
	logger := logging.LoggerFromContext(ctx).With("alias_id", aliasID)
	logger.InfoContext(ctx, "checking write access for product alias")

	// 1. Получение алиаса из БД
	alias, err := s.product.GetProductAliasByID(ctx, s.storage, aliasID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get product alias", "error", err)
		return fmt.Errorf("get product alias: %w", err)
	}

	// 2. Проверка прав на изменение через policy
	if err := policy.CanGroupAccessForModify(actor, alias); err != nil {
		logger.WarnContext(ctx, "write access denied", "error", err)
		return err
	}

	logger.InfoContext(ctx, "write access granted")
	return nil
}

// CreateProduct создаёт новый продукт в группе актора.
func (s *ServiceProduct) CreateProduct(ctx context.Context, actor policy.Actor, title string) (domain.ProductDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("title", title, "group_id", actor.GroupID)
	logger.InfoContext(ctx, "creating new product")

	var product domain.ProductDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		// Создание продукта с groupID актора
		product, err = s.product.CreateProduct(ctx, q, title, actor.GroupID)
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

// GetProductByID возвращает продукт по ID с проверкой прав на чтение.
func (s *ServiceProduct) GetProductByID(ctx context.Context, actor policy.Actor, productID int) (domain.ProductDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("product_id", productID)
	logger.InfoContext(ctx, "getting product by id")

	product, err := s.AccessReadProduct(ctx, actor, productID)
	if err != nil {
		return domain.ProductDetails{}, fmt.Errorf("access read product: %w", err)
	}

	logger.InfoContext(ctx, "product retrieved successfully")
	return product, nil
}

// UpdateProductByID обновляет продукт с проверкой прав на изменение.
func (s *ServiceProduct) UpdateProductByID(ctx context.Context, actor policy.Actor, productID int, updateProduct domain.ProductUpdate) (domain.ProductDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("product_id", productID)
	logger.InfoContext(ctx, "updating product")

	// 1. Проверка прав на запись
	if err := s.AccessWriteProduct(ctx, actor, productID); err != nil {
		return domain.ProductDetails{}, fmt.Errorf("access write product: %w", err)
	}

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

// DeleteProductByID удаляет продукт с проверкой прав на изменение.
func (s *ServiceProduct) DeleteProductByID(ctx context.Context, actor policy.Actor, productID int) error {
	logger := logging.LoggerFromContext(ctx).With("product_id", productID)
	logger.InfoContext(ctx, "deleting product")

	// 1. Проверка прав на запись
	if err := s.AccessWriteProduct(ctx, actor, productID); err != nil {
		return fmt.Errorf("access write product: %w", err)
	}

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

// ListProducts возвращает список продуктов, доступных в группе актора и общей группе.
func (s *ServiceProduct) ListProducts(ctx context.Context, actor policy.Actor) ([]domain.ProductDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("group_id", actor.GroupID)
	logger.InfoContext(ctx, "listing products for group")

	products, err := s.product.ListProducts(ctx, s.storage, []int{actor.GroupID, policy.CommonGroupID})
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

// CreateProductAlias создаёт алиас для продукта с проверкой прав на изменение продукта.
func (s *ServiceProduct) CreateProductAlias(ctx context.Context, actor policy.Actor, productID int, alias string) (domain.ProductAliasDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("product_id", productID, "alias", alias)
	logger.InfoContext(ctx, "creating product alias")

	// 1. Проверка прав на запись к продукту
	if err := s.AccessWriteProduct(ctx, actor, productID); err != nil {
		return domain.ProductAliasDetails{}, fmt.Errorf("access write product: %w", err)
	}

	var productAlias domain.ProductAliasDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		// 2. Создание алиаса в БД
		productAlias, err = s.product.CreateProductAlias(ctx, q, productID, alias, actor.GroupID)
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

// GetProductAliasByID возвращает алиас по ID с проверкой прав на чтение.
func (s *ServiceProduct) GetProductAliasByID(ctx context.Context, actor policy.Actor, aliasID int) (domain.ProductAliasDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("alias_id", aliasID)
	logger.InfoContext(ctx, "getting product alias by id")

	alias, err := s.AccessReadProductAlias(ctx, actor, aliasID)
	if err != nil {
		return domain.ProductAliasDetails{}, fmt.Errorf("access read product alias: %w", err)
	}

	logger.InfoContext(ctx, "product alias retrieved successfully")
	return alias, nil
}

// UpdateProductAliasByID обновляет алиас с проверкой прав на изменение алиаса.
func (s *ServiceProduct) UpdateProductAliasByID(ctx context.Context, actor policy.Actor, aliasID int, newAlias string) (domain.ProductAliasDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("alias_id", aliasID, "new_alias", newAlias)
	logger.InfoContext(ctx, "updating product alias")

	// 1. Проверка прав на запись к алиасу
	if err := s.AccessWriteProductAlias(ctx, actor, aliasID); err != nil {
		return domain.ProductAliasDetails{}, fmt.Errorf("access write product alias: %w", err)
	}

	var alias domain.ProductAliasDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		// 2. Обновление алиаса в БД
		alias, err = s.product.UpdateProductAliasByID(ctx, q, aliasID, domain.ProductAliasUpdate{Alias: &newAlias})
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

// DeleteProductAlias удаляет алиас с проверкой прав на изменение.
func (s *ServiceProduct) DeleteProductAlias(ctx context.Context, actor policy.Actor, aliasID int) error {
	logger := logging.LoggerFromContext(ctx).With("alias_id", aliasID)
	logger.InfoContext(ctx, "deleting product alias")

	// 1. Проверка прав на запись к алиасу
	if err := s.AccessWriteProductAlias(ctx, actor, aliasID); err != nil {
		return fmt.Errorf("access write product alias: %w", err)
	}

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

// ListProductAliases возвращает список алиасов продукта с проверкой прав на чтение продукта.
func (s *ServiceProduct) ListProductAliases(ctx context.Context, actor policy.Actor, productID int) ([]domain.ProductAliasDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("product_id", productID)
	logger.InfoContext(ctx, "listing product aliases for product")

	// 1. Проверка прав на чтение продукта (через AccessReadProduct)
	if _, err := s.AccessReadProduct(ctx, actor, productID); err != nil {
		return []domain.ProductAliasDetails{}, fmt.Errorf("access read product: %w", err)
	}

	// 2. Получение списка алиасов из БД
	aliases, err := s.product.ListProductAliases(ctx, s.storage, productID, []int{actor.GroupID, policy.CommonGroupID})
	if err != nil {
		logger.ErrorContext(ctx, "failed to list product aliases", "error", err)
		return []domain.ProductAliasDetails{}, fmt.Errorf("list product aliases: %w", err)
	}

	logger.InfoContext(ctx, "product aliases listed successfully", "count", len(aliases))
	return aliases, nil
}

// DeleteAllProductAliases удаляет все алиасы продукта с проверкой прав на изменение продукта.
func (s *ServiceProduct) DeleteAllProductAliases(ctx context.Context, actor policy.Actor, productID int) error {
	logger := logging.LoggerFromContext(ctx).With("product_id", productID)
	logger.InfoContext(ctx, "deleting all product aliases")

	// 1. Проверка прав на запись к продукту
	if err := s.AccessWriteProduct(ctx, actor, productID); err != nil {
		return fmt.Errorf("access write product: %w", err)
	}

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

// FindProductByAlias ищет название продукта по алиасу в группах актора и общей группе.
func (s *ServiceProduct) FindProductByAlias(ctx context.Context, actor policy.Actor, alias string) (string, error) {
	logger := logging.LoggerFromContext(ctx).With("alias", alias, "group_id", actor.GroupID)
	logger.InfoContext(ctx, "finding product by alias")

	// Поиск продукта по алиасу с фильтром по группам
	title, err := s.product.FindProductByAlias(ctx, s.storage, alias, []int{actor.GroupID, policy.CommonGroupID})
	if err != nil {
		logger.ErrorContext(ctx, "failed to find product by alias", "error", err)
		return "", fmt.Errorf("find product by alias: %w", err)
	}

	logger.InfoContext(ctx, "product found by alias", "title", title)
	return title, nil
}
