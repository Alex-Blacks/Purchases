package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
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
			err = fmt.Errorf("tx err: %v, commit err: %w", err, commitErr)
		}
	}()

	err = fn(tx)
	return err
}

func (s *ServiceAdminUnit) CreateUnit(ctx context.Context, name string, shortName string, groupID int) (domain.UnitDetails, error) {
	var unit domain.UnitDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		unit, err = s.unit.CreateUnit(ctx, q, name, groupID, shortName)
		return err
	}); err != nil {
		return domain.UnitDetails{}, err
	}
	return unit, nil
}

func (s *ServiceAdminUnit) GetUnit(ctx context.Context, unitID int) (domain.UnitDetails, error) {
	return s.unit.GetUnitByID(ctx, s.storage, unitID)
}

func (s *ServiceAdminUnit) UpdateUnit(ctx context.Context, unitID int, updateUnit domain.UnitUpdate) (domain.UnitDetails, error) {
	var unit domain.UnitDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		unit, err = s.unit.UpdateUnitByID(ctx, q, unitID, updateUnit)
		return err
	}); err != nil {
		return domain.UnitDetails{}, err
	}

	return unit, nil
}

func (s *ServiceAdminUnit) DeleteUnit(ctx context.Context, unitID int) error {
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.unit.DeleteUnitByID(ctx, q, unitID)
	})
}

func (s *ServiceAdminUnit) ListUnits(ctx context.Context) ([]domain.UnitDetails, error) {
	return s.unit.ListAdminUnits(ctx, s.storage)
}
