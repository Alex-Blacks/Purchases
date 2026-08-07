package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
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

func (s *ServiceUnit) WithTx(ctx context.Context, fn func(q domain.Querier) error) (err error) {
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

func (s *ServiceUnit) AccessReadUnit(ctx context.Context, actor policy.Actor, unitID int) (domain.UnitDetails, error) {
	unit, err := s.unit.GetUnitByID(ctx, s.storage, unitID)
	if err != nil {
		return domain.UnitDetails{}, err
	}
	if err := policy.CanGroupAccessForReading(actor, unit); err != nil {
		return domain.UnitDetails{}, err
	}
	return unit, nil
}

func (s *ServiceUnit) AccessWriteUnit(ctx context.Context, actor policy.Actor, unitID int) error {
	unit, err := s.unit.GetUnitByID(ctx, s.storage, unitID)
	if err != nil {
		return err
	}
	if err := policy.CanGroupAccessForModify(actor, unit); err != nil {
		return err
	}
	return nil
}

func (s *ServiceUnit) CreateUnit(ctx context.Context, actor policy.Actor, name string, shortName string) (domain.UnitDetails, error) {
	var unit domain.UnitDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {
		var err error
		unit, err = s.unit.CreateUnit(ctx, q, name, actor.GroupID, shortName)
		return err
	}); err != nil {
		return domain.UnitDetails{}, err
	}
	return unit, nil
}

func (s *ServiceUnit) GetUnit(ctx context.Context, actor policy.Actor, unitID int) (domain.UnitDetails, error) {
	unit, err := s.AccessReadUnit(ctx, actor, unitID)
	if err != nil {
		return domain.UnitDetails{}, err
	}
	return unit, nil
}

func (s *ServiceUnit) UpdateUnit(ctx context.Context, actor policy.Actor, unitID int, updateUnit domain.UnitUpdate) (domain.UnitDetails, error) {
	if err := s.AccessWriteUnit(ctx, actor, unitID); err != nil {
		return domain.UnitDetails{}, err
	}
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

func (s *ServiceUnit) DeleteUnit(ctx context.Context, actor policy.Actor, unitID int) error {
	if err := s.AccessWriteUnit(ctx, actor, unitID); err != nil {
		return err
	}
	return s.WithTx(ctx, func(q domain.Querier) error {
		return s.unit.DeleteUnitByID(ctx, q, unitID)
	})
}

func (s *ServiceUnit) ListUnits(ctx context.Context, actor policy.Actor) ([]domain.UnitDetails, error) {
	return s.unit.ListUnits(ctx, s.storage, []int{actor.GroupID, policy.CommonGroupID})
}
