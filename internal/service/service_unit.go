package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
)

type ServiceUnit struct {
	storage domain.Storage
	unit    domain.UnitRepository
}

func NewServiceUnit(st domain.Storage, unit domain.UnitRepository) *ServiceUnit {
	return &ServiceUnit{
		storage: st,
		unit:    unit,
	}
}

// withTx выполняет функцию в транзакции.
func (s *ServiceUnit) withTx(ctx context.Context, fn func(q domain.Querier) error) (err error) {
	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				logging.LoggerFromContext(ctx).ErrorContext(ctx, "rollback failed", "error", rollbackErr)
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

// accessReadUnit проверяет доступ к чтению единицы измерения и возвращает её.
func (s *ServiceUnit) accessReadUnit(ctx context.Context, q domain.Querier, actor policy.Actor, unitID int) (domain.UnitDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("unit_id", unitID)
	logger.InfoContext(ctx, "checking read access for unit")

	// 1. Получение единицы измерения из БД
	unit, err := s.unit.GetUnitByID(ctx, q, unitID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get unit", "error", err)
		return domain.UnitDetails{}, fmt.Errorf("get unit: %w", err)
	}

	// 2. Проверка прав на чтение через policy
	if err := policy.CanGroupAccessForReading(actor, unit); err != nil {
		logger.WarnContext(ctx, "read access denied", "error", err)
		return domain.UnitDetails{}, err
	}

	logger.InfoContext(ctx, "read access granted")
	return unit, nil
}

// accessWriteUnit проверяет доступ к изменению единицы измерения.
func (s *ServiceUnit) accessWriteUnit(ctx context.Context, q domain.Querier, actor policy.Actor, unitID int) error {
	logger := logging.LoggerFromContext(ctx).With("unit_id", unitID)
	logger.InfoContext(ctx, "checking write access for unit")

	// 1. Получение единицы измерения из БД
	unit, err := s.unit.GetUnitByID(ctx, q, unitID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get unit", "error", err)
		return fmt.Errorf("get unit: %w", err)
	}

	// 2. Проверка прав на изменение через policy
	if err := policy.CanGroupAccessForModify(actor, unit); err != nil {
		logger.WarnContext(ctx, "write access denied", "error", err)
		return err
	}

	logger.InfoContext(ctx, "write access granted")
	return nil
}

// CreateUnit создаёт новую единицу измерения в группе актора.
func (s *ServiceUnit) CreateUnit(ctx context.Context, actor policy.Actor, name string, shortName string, groupID *int) (domain.UnitDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("name", name, "short_name", shortName)
	if groupID != nil {
		logger = logger.With("requested_group_id", groupID)
	}
	logger.InfoContext(ctx, "creating new unit")

	if strings.TrimSpace(name) == "" || strings.TrimSpace(shortName) == "" {
		return domain.UnitDetails{}, domain.ErrEmptyName
	}
	var unit domain.UnitDetails
	if err := s.withTx(ctx, func(q domain.Querier) error {
		var err error
		if actor.HasRole(policy.RoleAdmin) {
			// Создание единицы измерения с groupID которое передал админ
			if groupID == nil || *groupID < 1 {
				return domain.ErrGroupIDRequired
			}
			unit, err = s.unit.CreateUnit(ctx, q, name, *groupID, shortName)
		} else {
			// Создание единицы измерения с groupID актора
			if groupID != nil {
				return domain.ErrGroupIDNotAllowed
			}
			unit, err = s.unit.CreateUnit(ctx, q, name, actor.GroupID, shortName)
		}
		if err != nil {
			logger.ErrorContext(ctx, "failed to create unit", "error", err)
			return fmt.Errorf("create unit: %w", err)
		}
		return nil
	}); err != nil {
		return domain.UnitDetails{}, err
	}

	logger.InfoContext(ctx, "unit created successfully", "unit_id", unit.ID)
	return unit, nil
}

// GetUnit возвращает единицу измерения по ID с проверкой прав на чтение.
func (s *ServiceUnit) GetUnit(ctx context.Context, actor policy.Actor, unitID int) (domain.UnitDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("unit_id", unitID)
	logger.InfoContext(ctx, "getting unit by id")

	if unitID < 1 {
		return domain.UnitDetails{}, domain.ErrInvalidInput
	}
	unit, err := s.accessReadUnit(ctx, s.storage, actor, unitID)
	if err != nil {
		return domain.UnitDetails{}, fmt.Errorf("access read unit: %w", err)
	}

	logger.InfoContext(ctx, "unit retrieved successfully")
	return unit, nil
}

// UpdateUnit обновляет единицу измерения с проверкой прав на изменение.
func (s *ServiceUnit) UpdateUnit(ctx context.Context, actor policy.Actor, unitID int, updateUnit domain.UnitUpdate) (domain.UnitDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("unit_id", unitID)
	logger.InfoContext(ctx, "updating unit")

	if unitID < 1 {
		return domain.UnitDetails{}, domain.ErrInvalidInput
	}
	if updateUnit.Name != nil && strings.TrimSpace(*updateUnit.Name) == "" {
		return domain.UnitDetails{}, domain.ErrEmptyName
	}
	if updateUnit.ShortName != nil && strings.TrimSpace(*updateUnit.ShortName) == "" {
		return domain.UnitDetails{}, domain.ErrEmptyName
	}
	var unit domain.UnitDetails
	if err := s.withTx(ctx, func(q domain.Querier) error {
		var err error
		// 1. Проверка прав на запись
		if err := s.accessWriteUnit(ctx, q, actor, unitID); err != nil {
			return fmt.Errorf("access write unit: %w", err)
		}

		// 2. Обновление единицы измерения в БД
		unit, err = s.unit.UpdateUnitByID(ctx, q, unitID, updateUnit)
		if err != nil {
			logger.ErrorContext(ctx, "failed to update unit", "error", err)
			return fmt.Errorf("update unit: %w", err)
		}
		return nil
	}); err != nil {
		return domain.UnitDetails{}, err
	}

	logger.InfoContext(ctx, "unit updated successfully")
	return unit, nil
}

// DeleteUnit удаляет единицу измерения с проверкой прав на изменение.
func (s *ServiceUnit) DeleteUnit(ctx context.Context, actor policy.Actor, unitID int) error {
	logger := logging.LoggerFromContext(ctx).With("unit_id", unitID)
	logger.InfoContext(ctx, "deleting unit")

	if unitID < 1 {
		return domain.ErrInvalidInput
	}
	if err := s.withTx(ctx, func(q domain.Querier) error {
		// 1. Проверка прав на запись
		if err := s.accessWriteUnit(ctx, q, actor, unitID); err != nil {
			return fmt.Errorf("access write unit: %w", err)
		}

		// 2. Удаление единицы измерения в транзакции
		if err := s.unit.DeleteUnitByID(ctx, q, unitID); err != nil {
			logger.ErrorContext(ctx, "failed to delete unit", "error", err)
			return fmt.Errorf("delete unit: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	logger.InfoContext(ctx, "unit deleted successfully")
	return nil
}

// ListUnits возвращает список единиц измерения, доступных в группе актора и общей группе.
func (s *ServiceUnit) ListUnits(ctx context.Context, actor policy.Actor) ([]domain.UnitDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("group_id", actor.GroupID)
	logger.InfoContext(ctx, "listing units for group")

	// Получение списка единиц измерения из БД (без транзакции)
	units, err := s.unit.ListUnits(ctx, s.storage, []int{actor.GroupID, policy.CommonGroupID})
	if err != nil {
		logger.ErrorContext(ctx, "failed to list units", "error", err)
		return []domain.UnitDetails{}, fmt.Errorf("list units: %w", err)
	}

	logger.InfoContext(ctx, "units listed successfully", "count", len(units))
	return units, nil
}

// ListUnits возвращает список всех единиц измерения. Доступно только администраторам.
func (s *ServiceUnit) ListAllUnits(ctx context.Context, actor policy.Actor) ([]domain.UnitDetails, error) {
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
