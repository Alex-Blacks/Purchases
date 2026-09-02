package service

import (
	"context"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/policy"
)

type ServiceUnit struct {
	*GenericService[
		domain.UnitDetails,
		domain.UnitCreate,
		domain.UnitUpdate,
		domain.UnitListFilter,
		domain.UnitRepository]
}

func NewServiceUnit(st domain.Storage, repo domain.UnitRepository, history domain.ChangeHistoryRepository) *ServiceUnit {
	return &ServiceUnit{
		GenericService: &GenericService[domain.UnitDetails, domain.UnitCreate, domain.UnitUpdate, domain.UnitListFilter, domain.UnitRepository]{
			BaseService: &BaseService{storage: st},
			repo:        repo,
			history:     history,
			entityType:  domain.HistoryEntityUnit,
		},
	}
}

// Create создаёт новую единицу измерения в указанной группе или группе актора.
func (s *ServiceUnit) Create(ctx context.Context, actor policy.Actor, params domain.UnitCreate, groupID *int) (domain.UnitDetails, error) {
	// 1. Валидация
	if strings.TrimSpace(params.Name) == "" {
		return domain.UnitDetails{}, domain.ErrEmptyName
	}
	if strings.TrimSpace(params.ShortName) == "" {
		return domain.UnitDetails{}, domain.ErrEmptyName
	}

	return s.GenericService.Create(ctx, actor, params, groupID)
}

// Update обновляет единицу измерения.
func (s *ServiceUnit) Update(ctx context.Context, actor policy.Actor, id int, updates domain.UnitUpdate) (domain.UnitDetails, error) {
	// 1. Валидация
	if updates.Name != nil && strings.TrimSpace(*updates.Name) == "" {
		return domain.UnitDetails{}, domain.ErrEmptyName
	}
	if updates.ShortName != nil && strings.TrimSpace(*updates.ShortName) == "" {
		return domain.UnitDetails{}, domain.ErrEmptyName
	}

	return s.GenericService.Update(ctx, actor, id, updates)
}

// List возвращает список единиц измерения, с фильтрацией.
func (s *ServiceUnit) List(ctx context.Context, actor policy.Actor, filter domain.UnitListFilter) ([]domain.UnitDetails, error) {
	// 1. Валидация фильтра
	if err := prepareUnitFilter(actor, &filter); err != nil {
		return nil, err
	}

	return s.GenericService.List(ctx, actor, filter)
}

// Count возвращает количество единиц измерения, с фильтрацией.
func (s *ServiceUnit) Count(ctx context.Context, actor policy.Actor, filter domain.UnitListFilter) (int, error) {
	// 1. Валидация фильтра
	if err := prepareUnitFilter(actor, &filter); err != nil {
		return 0, err
	}

	return s.GenericService.Count(ctx, actor, filter)
}
