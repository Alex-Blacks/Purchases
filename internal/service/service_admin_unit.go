package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
)

type ServiceAdminUnit struct {
	storage domain.Storage
	unit    domain.UnitRepository
}

func NewServiceAdminUnit(st domain.Storage, unit domain.UnitRepository) *ServiceAdminUnit {
	return &ServiceAdminUnit{
		storage: st,
		unit:    unit,
	}
}

// WithTx выполняет функцию в транзакции.
func (s *ServiceAdminUnit) WithTx(ctx context.Context, fn func(q domain.Querier) error) (err error) {
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

// CreateUnit создаёт новую единицу измерения. Доступно только администраторам.
func (s *ServiceAdminUnit) CreateUnit(ctx context.Context, actor policy.Actor, name string, shortName string, groupID int) (domain.UnitDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.UnitDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("name", name, "short_name", shortName, "group_id", groupID)
	logger.InfoContext(ctx, "creating new unit")

	var unit domain.UnitDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		// 2. Создание единицы измерения в БД
		unit, err = s.unit.CreateUnit(ctx, q, name, groupID, shortName)
		if err != nil {
			logger.ErrorContext(ctx, "failed to create unit", "error", err)
			return fmt.Errorf("create unit: %w", err)
		}
		return nil
	}); err != nil {
		return domain.UnitDetails{}, fmt.Errorf("create unit: %w", err)
	}

	logger.InfoContext(ctx, "unit created successfully", "unit_id", unit.ID)
	return unit, nil
}

// GetUnit возвращает единицу измерения по ID. Доступно только администраторам.
func (s *ServiceAdminUnit) GetUnit(ctx context.Context, actor policy.Actor, unitID int) (domain.UnitDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.UnitDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("unit_id", unitID)
	logger.InfoContext(ctx, "getting unit by id")

	// 2. Получение единицы измерения из БД (без транзакции)
	unit, err := s.unit.GetUnitByID(ctx, s.storage, unitID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get unit", "error", err)
		return domain.UnitDetails{}, fmt.Errorf("get unit: %w", err)
	}

	logger.InfoContext(ctx, "unit retrieved successfully")
	return unit, nil
}

// UpdateUnit обновляет единицу измерения. Доступно только администраторам.
func (s *ServiceAdminUnit) UpdateUnit(ctx context.Context, actor policy.Actor, unitID int, updateUnit domain.UnitUpdate) (domain.UnitDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return domain.UnitDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("unit_id", unitID)
	logger.InfoContext(ctx, "updating unit")

	var unit domain.UnitDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		// 2. Обновление единицы измерения в БД
		unit, err = s.unit.UpdateUnitByID(ctx, q, unitID, updateUnit)
		if err != nil {
			logger.ErrorContext(ctx, "failed to update unit", "error", err)
			return fmt.Errorf("update unit: %w", err)
		}
		return nil
	}); err != nil {
		return domain.UnitDetails{}, fmt.Errorf("update unit: %w", err)
	}

	logger.InfoContext(ctx, "unit updated successfully")
	return unit, nil
}

// DeleteUnit удаляет единицу измерения по ID. Доступно только администраторам.
func (s *ServiceAdminUnit) DeleteUnit(ctx context.Context, actor policy.Actor, unitID int) error {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx).With("unit_id", unitID)
	logger.InfoContext(ctx, "deleting unit")

	// 2. Удаление единицы измерения в транзакции
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		if err := s.unit.DeleteUnitByID(ctx, q, unitID); err != nil {
			logger.ErrorContext(ctx, "failed to delete unit", "error", err)
			return fmt.Errorf("delete unit: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("delete unit: %w", err)
	}

	logger.InfoContext(ctx, "unit deleted successfully")
	return nil
}

// ListUnits возвращает список всех единиц измерения. Доступно только администраторам.
func (s *ServiceAdminUnit) ListUnits(ctx context.Context, actor policy.Actor) ([]domain.UnitDetails, error) {
	// 1. Проверка прав
	if !actor.HasRole(policy.RoleAdmin) {
		return []domain.UnitDetails{}, policy.ErrForbidden
	}

	logger := logging.LoggerFromContext(ctx)
	logger.InfoContext(ctx, "listing units")

	// 2. Получение списка единиц измерения из БД (без транзакции)
	units, err := s.unit.ListAdminUnits(ctx, s.storage)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list units", "error", err)
		return []domain.UnitDetails{}, fmt.Errorf("list units: %w", err)
	}

	logger.InfoContext(ctx, "units listed successfully", "count", len(units))
	return units, nil
}
