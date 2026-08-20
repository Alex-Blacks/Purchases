package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
)

type ServiceProduct struct {
	*GenericService[domain.ProductDetails, domain.ProductRepository]
}

func NewServiceProduct(st domain.Storage, repo domain.ProductRepository) *ServiceProduct {
	return &ServiceProduct{
		GenericService: &GenericService[domain.ProductDetails, domain.ProductRepository]{
			BaseService: &BaseService{storage: st},
			repo:        repo,
		},
	}
}

// ---------------------------------------------------------------------------------
// --------------------------------------ALIAS--------------------------------------
// ---------------------------------------------------------------------------------

type ServiceProductAlias struct {
	*BaseService
	repo domain.ProductAliasRepository
}

func NewServiceProductAlias(st domain.Storage, repo domain.ProductAliasRepository) *ServiceProductAlias {
	return &ServiceProductAlias{
		BaseService: &BaseService{storage: st},
		repo:        repo,
	}
}

func (s *ServiceProductAlias) getEntity(ctx context.Context, q domain.Querier, id int) (domain.GroupedEntity, error) {
	return s.repo.GetByID(ctx, q, id)
}

// Create создаёт алиас для продукта с проверкой прав на изменение продукта.
func (s *ServiceProductAlias) Create(ctx context.Context, actor policy.Actor, productID int, alias string, groupID *int) (domain.ProductAliasDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("product_id", productID, "alias", alias)
	logger.InfoContext(ctx, "creating product alias")

	if productID < 1 {
		return domain.ProductAliasDetails{}, domain.ErrInvalidInput
	}
	if strings.TrimSpace(alias) == "" {
		return domain.ProductAliasDetails{}, domain.ErrEmptyName
	}

	targetGroup, err := s.resolveGroupID(actor, groupID)
	if err != nil {
		return domain.ProductAliasDetails{}, err
	}

	var productAlias domain.ProductAliasDetails
	if err := s.withTx(ctx, func(q domain.Querier) error {
		var err error
		// 1. Проверка прав на запись к продукту
		if err := s.accessWrite(ctx, q, actor, productID, s.getEntity); err != nil {
			return fmt.Errorf("access write product: %w", err)
		}

		// 2. Создание алиаса в БД
		productAlias, err = s.repo.Create(ctx, q, productID, alias, targetGroup)
		if err != nil {
			logger.ErrorContext(ctx, "failed to create product alias", "error", err)
			return fmt.Errorf("create product alias: %w", err)
		}
		return nil
	}); err != nil {
		return domain.ProductAliasDetails{}, err
	}

	logger.InfoContext(ctx, "product alias created successfully", "alias_id", productAlias.ID)
	return productAlias, nil
}

// GetByID возвращает алиас по ID с проверкой прав на чтение.
func (s *ServiceProductAlias) GetByID(ctx context.Context, actor policy.Actor, aliasID int) (domain.ProductAliasDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("alias_id", aliasID)
	logger.InfoContext(ctx, "getting product alias by id")

	if aliasID < 1 {
		return domain.ProductAliasDetails{}, domain.ErrInvalidInput
	}
	alias, err := s.accessRead(ctx, s.storage, actor, aliasID, s.getEntity)
	if err != nil {
		return domain.ProductAliasDetails{}, fmt.Errorf("access read product alias: %w", err)
	}

	logger.InfoContext(ctx, "product alias retrieved successfully")
	return alias.(domain.ProductAliasDetails), nil
}

// UpdateByID обновляет алиас с проверкой прав на изменение алиаса.
func (s *ServiceProductAlias) UpdateByID(ctx context.Context, actor policy.Actor, aliasID int, newAlias string) (domain.ProductAliasDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("alias_id", aliasID, "new_alias", newAlias)
	logger.InfoContext(ctx, "updating product alias")

	if aliasID < 1 {
		return domain.ProductAliasDetails{}, domain.ErrInvalidInput
	}

	if strings.TrimSpace(newAlias) == "" {
		return domain.ProductAliasDetails{}, domain.ErrEmptyName
	}

	var alias domain.ProductAliasDetails
	if err := s.withTx(ctx, func(q domain.Querier) error {
		var err error
		// 1. Проверка прав на запись к алиасу
		if err := s.accessWrite(ctx, q, actor, aliasID, s.getEntity); err != nil {
			return fmt.Errorf("access write product alias: %w", err)
		}

		// 2. Обновление алиаса в БД
		alias, err = s.repo.UpdateByID(ctx, q, aliasID, domain.ProductAliasUpdate{Alias: &newAlias})
		if err != nil {
			logger.ErrorContext(ctx, "failed to update product alias", "error", err)
			return fmt.Errorf("update product alias: %w", err)
		}
		return nil
	}); err != nil {
		return domain.ProductAliasDetails{}, err
	}

	logger.InfoContext(ctx, "product alias updated successfully")
	return alias, nil
}

// DeleteByID удаляет алиас с проверкой прав на изменение.
func (s *ServiceProductAlias) DeleteByID(ctx context.Context, actor policy.Actor, aliasID int) error {
	logger := logging.LoggerFromContext(ctx).With("alias_id", aliasID)
	logger.InfoContext(ctx, "deleting product alias")

	if aliasID < 1 {
		return domain.ErrInvalidInput
	}

	if err := s.withTx(ctx, func(q domain.Querier) error {
		// 1. Проверка прав на запись к алиасу
		if err := s.accessWrite(ctx, q, actor, aliasID, s.getEntity); err != nil {
			return fmt.Errorf("access write product alias: %w", err)
		}

		// 2. Удаление алиаса в транзакции
		if err := s.repo.DeleteByID(ctx, q, aliasID); err != nil {
			logger.ErrorContext(ctx, "failed to delete product alias", "error", err)
			return fmt.Errorf("delete product alias: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	logger.InfoContext(ctx, "product alias deleted successfully")
	return nil
}

// List возвращает список алиасов продукта с проверкой прав на чтение продукта.
func (s *ServiceProductAlias) List(ctx context.Context, actor policy.Actor, productID int) ([]domain.ProductAliasDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("product_id", productID)
	logger.InfoContext(ctx, "listing product aliases for product")

	if productID < 1 {
		return []domain.ProductAliasDetails{}, domain.ErrInvalidInput
	}

	// 1. Проверка прав на чтение продукта (через accessRead)
	if _, err := s.accessRead(ctx, s.storage, actor, productID, s.getEntity); err != nil {
		return []domain.ProductAliasDetails{}, fmt.Errorf("access read product: %w", err)
	}

	// 2. Получение списка алиасов из БД
	aliases, err := s.repo.List(ctx, s.storage, productID, []int{actor.GroupID, policy.CommonGroupID})
	if err != nil {
		logger.ErrorContext(ctx, "failed to list product aliases", "error", err)
		return []domain.ProductAliasDetails{}, fmt.Errorf("list product aliases: %w", err)
	}

	logger.InfoContext(ctx, "product aliases listed successfully", "count", len(aliases))
	return aliases, nil
}

// ListAll возвращает список алиасов продукта с проверкой прав на чтение продукта.
func (s *ServiceProductAlias) ListAll(ctx context.Context, actor policy.Actor, productID int) ([]domain.ProductAliasDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("product_id", productID)
	logger.InfoContext(ctx, "listing product aliases for product")

	if productID < 1 {
		return []domain.ProductAliasDetails{}, domain.ErrInvalidInput
	}

	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return []domain.ProductAliasDetails{}, policy.ErrForbidden
	}

	// 2. Получение списка алиасов из БД
	aliases, err := s.repo.ListAll(ctx, s.storage, productID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list product aliases", "error", err)
		return []domain.ProductAliasDetails{}, fmt.Errorf("list product aliases: %w", err)
	}

	logger.InfoContext(ctx, "product aliases listed successfully", "count", len(aliases))
	return aliases, nil
}

// DeleteAllProductAliases удаляет все алиасы продукта с проверкой прав на изменение продукта.
func (s *ServiceProductAlias) DeleteAllProductAliases(ctx context.Context, actor policy.Actor, productID int) error {
	logger := logging.LoggerFromContext(ctx).With("product_id", productID)
	logger.InfoContext(ctx, "deleting all product aliases")

	if productID < 1 {
		return domain.ErrInvalidInput
	}

	if err := s.withTx(ctx, func(q domain.Querier) error {
		// 1. Проверка прав на запись к продукту
		if err := s.accessWrite(ctx, q, actor, productID, s.getEntity); err != nil {
			return fmt.Errorf("access write product: %w", err)
		}

		// 2. Удаление всех алиасов в транзакции
		if err := s.repo.DeleteAllProductAliases(ctx, q, productID); err != nil {
			logger.ErrorContext(ctx, "failed to delete all product aliases", "error", err)
			return fmt.Errorf("delete all product aliases: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	logger.InfoContext(ctx, "all product aliases deleted successfully")
	return nil
}

// FindProductByAlias ищет название продукта по алиасу в группах актора и общей группе.
func (s *ServiceProductAlias) FindProductByAlias(ctx context.Context, actor policy.Actor, alias string) (string, error) {
	logger := logging.LoggerFromContext(ctx).With("alias", alias, "group_id", actor.GroupID)
	logger.InfoContext(ctx, "finding product by alias")

	if strings.TrimSpace(alias) == "" {
		return "", domain.ErrEmptyName
	}

	// Поиск продукта по алиасу с фильтром по группам
	title, err := s.repo.FindProductByAlias(ctx, s.storage, alias, []int{actor.GroupID, policy.CommonGroupID})
	if err != nil {
		logger.ErrorContext(ctx, "failed to find product by alias", "error", err)
		return "", fmt.Errorf("find product by alias: %w", err)
	}

	logger.InfoContext(ctx, "product found by alias", "title", title)
	return title, nil
}

// FindAllProductByAlias ищет название продукта по алиасу в группах актора и общей группе.
func (s *ServiceProductAlias) FindAllProductByAlias(ctx context.Context, actor policy.Actor, alias string) (string, error) {
	logger := logging.LoggerFromContext(ctx).With("alias", alias)
	logger.InfoContext(ctx, "finding all product by alias")

	if strings.TrimSpace(alias) == "" {
		return "", domain.ErrEmptyName
	}

	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return "", policy.ErrForbidden
	}

	// 2. Поиск продукта по алиасу
	title, err := s.repo.FindAllProductByAlias(ctx, s.storage, alias)
	if err != nil {
		logger.ErrorContext(ctx, "failed to find product by alias", "error", err)
		return "", fmt.Errorf("find product by alias: %w", err)
	}

	logger.InfoContext(ctx, "product found by alias", "title", title)
	return title, nil
}
