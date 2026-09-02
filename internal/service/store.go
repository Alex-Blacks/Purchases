package service

import (
	"context"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/policy"
)

type ServiceStore struct {
	*GenericService[
		domain.StoreDetails,
		domain.StoreCreate,
		domain.StoreUpdate,
		domain.StoreListFilter,
		domain.StoreRepository]
}

func NewServiceStore(st domain.Storage, repo domain.StoreRepository, history domain.ChangeHistoryRepository) *ServiceStore {
	return &ServiceStore{
		GenericService: &GenericService[domain.StoreDetails, domain.StoreCreate, domain.StoreUpdate, domain.StoreListFilter, domain.StoreRepository]{
			BaseService: &BaseService{storage: st},
			repo:        repo,
			history:     history,
			entityType:  domain.HistoryEntityStore,
		},
	}
}

// Create создаёт новый магазин в указанной группе или группе актора.
func (s *ServiceStore) Create(ctx context.Context, actor policy.Actor, params domain.StoreCreate, groupID *int) (domain.StoreDetails, error) {
	// 1. Валидация
	if strings.TrimSpace(params.Name) == "" {
		return domain.StoreDetails{}, domain.ErrEmptyName
	}

	return s.GenericService.Create(ctx, actor, params, groupID)
}

// Update обновляет магазин.
func (s *ServiceStore) Update(ctx context.Context, actor policy.Actor, id int, updates domain.StoreUpdate) (domain.StoreDetails, error) {
	// 1. Валидация
	if updates.Name != nil && strings.TrimSpace(*updates.Name) == "" {
		return domain.StoreDetails{}, domain.ErrEmptyName
	}

	return s.GenericService.Update(ctx, actor, id, updates)
}

// List возвращает список магазинов, с фильтрацией.
func (s *ServiceStore) List(ctx context.Context, actor policy.Actor, filter domain.StoreListFilter) ([]domain.StoreDetails, error) {
	// 1. Валидация фильтра
	if err := prepareStoreFilter(actor, &filter); err != nil {
		return nil, err
	}

	return s.GenericService.List(ctx, actor, filter)
}

// Count возвращает количество магазинов, с фильтрацией.
func (s *ServiceStore) Count(ctx context.Context, actor policy.Actor, filter domain.StoreListFilter) (int, error) {
	// 1. Валидация фильтра
	if err := prepareStoreFilter(actor, &filter); err != nil {
		return 0, err
	}

	return s.GenericService.Count(ctx, actor, filter)
}
