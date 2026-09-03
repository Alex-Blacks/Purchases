package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
)

type ServiceProduct struct {
	*GenericService[
		domain.ProductDetails,
		domain.ProductCreate,
		domain.ProductUpdate,
		domain.ProductListFilter,
		domain.ProductRepository]
}

func NewServiceProduct(st domain.Storage, repo domain.ProductRepository, history domain.ChangeHistoryRepository) *ServiceProduct {
	return &ServiceProduct{
		GenericService: &GenericService[domain.ProductDetails, domain.ProductCreate, domain.ProductUpdate, domain.ProductListFilter, domain.ProductRepository]{
			BaseService: &BaseService{storage: st},
			repo:        repo,
			history:     history,
			entityType:  domain.HistoryEntityProduct,
		},
	}
}

// Create создаёт новый продукт в указанной группе или группе актора.
func (s *ServiceProduct) Create(ctx context.Context, actor policy.Actor, params domain.ProductCreate, groupID *int) (domain.ProductDetails, error) {
	// 1. Валидация
	if strings.TrimSpace(params.Title) == "" {
		return domain.ProductDetails{}, domain.ErrEmptyName
	}

	return s.GenericService.Create(ctx, actor, params, groupID)
}

// Update обновляет продукт.
func (s *ServiceProduct) Update(ctx context.Context, actor policy.Actor, id int, updates domain.ProductUpdate) (domain.ProductDetails, error) {
	// 1. Валидация
	if updates.Title != nil && strings.TrimSpace(*updates.Title) == "" {
		return domain.ProductDetails{}, domain.ErrEmptyName
	}

	return s.GenericService.Update(ctx, actor, id, updates)
}

// List возвращает список продуктов, с фильтрацией.
func (s *ServiceProduct) List(ctx context.Context, actor policy.Actor, filter domain.ProductListFilter) ([]domain.ProductDetails, error) {
	// 1. Валидация фильтра
	if err := prepareProductFilter(actor, &filter); err != nil {
		return nil, err
	}

	return s.GenericService.List(ctx, actor, filter)
}

// Count возвращает количество продуктов, с фильтрацией.
func (s *ServiceProduct) Count(ctx context.Context, actor policy.Actor, filter domain.ProductListFilter) (int, error) {
	// 1. Валидация фильтра
	if err := prepareProductFilter(actor, &filter); err != nil {
		return 0, err
	}

	return s.GenericService.Count(ctx, actor, filter)
}

// ---------------------------------------------------------------------------------
// --------------------------------------ALIAS--------------------------------------
// ---------------------------------------------------------------------------------

type ServiceProductAlias struct {
	*BaseService
	repo        domain.ProductAliasRepository
	productRepo domain.ProductRepository
	history     domain.ChangeHistoryRepository
}

func NewServiceProductAlias(st domain.Storage, repo domain.ProductAliasRepository, productRepo domain.ProductRepository, history domain.ChangeHistoryRepository) *ServiceProductAlias {
	return &ServiceProductAlias{
		BaseService: &BaseService{storage: st},
		repo:        repo,
		productRepo: productRepo,
		history:     history,
	}
}

func (s *ServiceProductAlias) getProductEntity(ctx context.Context, q domain.Querier, id int) (domain.GroupedEntity, error) {
	return s.productRepo.GetByID(ctx, q, id)
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
		if err := s.accessWrite(ctx, q, actor, productID, s.getProductEntity); err != nil {
			return fmt.Errorf("access write product: %w", err)
		}

		// 2. Создание алиаса в БД
		productAlias, err = s.repo.Create(ctx, q, productID, alias, targetGroup)
		if err != nil {
			logger.ErrorContext(ctx, "failed to create product alias", "error", err)
			return fmt.Errorf("create product alias: %w", err)
		}

		// 3. Запись истории создания
		newData, err := json.Marshal(productAlias)
		if err != nil {
			logger.ErrorContext(ctx, "failed marshaling to json", "error", err)
			return fmt.Errorf("marshaling to json: %w", err)
		}
		if err := s.history.Insert(ctx, q, actor.GroupID, actor.UserID, domain.HistoryEntityProductAlias, productAlias.ID, domain.HistoryActionCreate, nil, newData); err != nil {
			logger.ErrorContext(ctx, "failed to insert history", "error", err)
			return fmt.Errorf("insert history: %w", err)
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
	alias, err := s.accessRead(ctx, s.storage, actor, aliasID, s.getProductEntity)
	if err != nil {
		return domain.ProductAliasDetails{}, fmt.Errorf("access read product alias: %w", err)
	}

	result, ok := alias.(domain.ProductAliasDetails)
	if !ok {
		return domain.ProductAliasDetails{}, fmt.Errorf("unexpected entity type")
	}
	logger.InfoContext(ctx, "product alias retrieved successfully")
	return result, nil
}

// UpdateByID обновляет алиас с проверкой прав на изменение алиаса.
func (s *ServiceProductAlias) UpdateByID(ctx context.Context, actor policy.Actor, aliasID int, updates domain.ProductAliasUpdate) (domain.ProductAliasDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("alias_id", aliasID, "updates", updates)
	logger.InfoContext(ctx, "updating product alias")

	if aliasID < 1 {
		return domain.ProductAliasDetails{}, domain.ErrInvalidInput
	}

	if updates.Alias != nil && strings.TrimSpace(*updates.Alias) == "" {
		return domain.ProductAliasDetails{}, domain.ErrEmptyName
	}

	var alias domain.ProductAliasDetails
	if err := s.withTx(ctx, func(q domain.Querier) error {
		var err error
		// 1. Проверка прав на запись к алиасу
		if err := s.accessWrite(ctx, q, actor, aliasID, s.getProductEntity); err != nil {
			return fmt.Errorf("access write product alias: %w", err)
		}

		// 2. Получение старой версии для истории
		oldAlias, err := s.repo.GetByID(ctx, q, aliasID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to get alias for history", "error", err)
			return fmt.Errorf("get alias for history: %w", err)
		}
		oldData, err := json.Marshal(oldAlias)
		if err != nil {
			logger.ErrorContext(ctx, "failed marshaling to json", "error", err)
			return fmt.Errorf("marshaling to json: %w", err)
		}

		// 3. Обновление алиаса в БД
		alias, err = s.repo.UpdateByID(ctx, q, aliasID, updates)
		if err != nil {
			logger.ErrorContext(ctx, "failed to update product alias", "error", err)
			return fmt.Errorf("update product alias: %w", err)
		}

		// 4. Запись истории обновления
		newData, err := json.Marshal(alias)
		if err != nil {
			logger.ErrorContext(ctx, "failed marshaling to json", "error", err)
			return fmt.Errorf("marshaling to json: %w", err)
		}
		if err := s.history.Insert(ctx, q, actor.GroupID, actor.UserID, domain.HistoryEntityProductAlias, alias.ID, domain.HistoryActionUpdate, oldData, newData); err != nil {
			logger.ErrorContext(ctx, "failed to insert history", "error", err)
			return fmt.Errorf("insert history: %w", err)
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
		if err := s.accessWrite(ctx, q, actor, aliasID, s.getProductEntity); err != nil {
			return fmt.Errorf("access write product alias: %w", err)
		}

		// 2. Получение старой версии для истории
		oldAlias, err := s.repo.GetByID(ctx, q, aliasID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to get alias for history", "error", err)
			return fmt.Errorf("get alias for history: %w", err)
		}
		oldData, err := json.Marshal(oldAlias)
		if err != nil {
			logger.ErrorContext(ctx, "failed marshaling to json", "error", err)
			return fmt.Errorf("marshaling to json: %w", err)
		}

		// 3. Удаление алиаса в транзакции
		if err := s.repo.DeleteByID(ctx, q, aliasID); err != nil {
			logger.ErrorContext(ctx, "failed to delete product alias", "error", err)
			return fmt.Errorf("delete product alias: %w", err)
		}

		// 4. Запись истории удаления
		if err := s.history.Insert(ctx, q, actor.GroupID, actor.UserID, domain.HistoryEntityProductAlias, oldAlias.ID, domain.HistoryActionDelete, oldData, nil); err != nil {
			logger.ErrorContext(ctx, "failed to insert history", "error", err)
			return fmt.Errorf("insert history: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	logger.InfoContext(ctx, "product alias deleted successfully")
	return nil
}

// List возвращает список алиасов продукта с проверкой прав на чтение продукта.
func (s *ServiceProductAlias) List(ctx context.Context, actor policy.Actor, filter domain.ProductAliasListFilter) ([]domain.ProductAliasDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("product_id", filter.ProductID, "filter", filter)
	logger.InfoContext(ctx, "listing product aliases for product")

	if err := prepareProductAliasFilter(actor, &filter); err != nil {
		return nil, err
	}

	// 2. Получение списка алиасов из БД
	aliases, err := s.repo.List(ctx, s.storage, filter)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list product aliases", "error", err)
		return nil, fmt.Errorf("list product aliases: %w", err)
	}

	logger.InfoContext(ctx, "product aliases listed successfully", "count", len(aliases))
	return aliases, nil
}

// Count возвращает количество алиасов продукта с проверкой прав на чтение продукта.
func (s *ServiceProductAlias) Count(ctx context.Context, actor policy.Actor, filter domain.ProductAliasListFilter) (int, error) {
	logger := logging.LoggerFromContext(ctx).With("product_id", filter.ProductID, "filter", filter)
	logger.InfoContext(ctx, "counting product aliases for product")

	// 2. Получение списка алиасов из БД
	count, err := s.repo.Count(ctx, s.storage, filter)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list product aliases", "error", err)
		return 0, fmt.Errorf("list product aliases: %w", err)
	}

	logger.InfoContext(ctx, "product aliases listed successfully", "count", count)
	return count, nil
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
		if err := s.accessWrite(ctx, q, actor, productID, s.getProductEntity); err != nil {
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
func (s *ServiceProductAlias) FindProductByAlias(ctx context.Context, actor policy.Actor, alias string) (domain.ProductDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("alias", alias, "group_id", actor.GroupID)
	logger.InfoContext(ctx, "finding product by alias")

	if strings.TrimSpace(alias) == "" {
		return domain.ProductDetails{}, domain.ErrEmptyName
	}

	var groupIDs []int
	if !actor.HasRole(domain.RoleAdmin) {
		groupIDs = []int{actor.GroupID, policy.CommonGroupID}
	}

	// Поиск продукта по алиасу с фильтром по группам
	product, err := s.repo.FindProductByAlias(ctx, s.storage, alias, groupIDs)
	if err != nil {
		logger.ErrorContext(ctx, "failed to find product by alias", "error", err)
		return domain.ProductDetails{}, fmt.Errorf("find product by alias: %w", err)
	}

	logger.InfoContext(ctx, "product found by alias", "product_id", product.ID, "title", product.Title)
	return product, nil
}
