package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
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

func (s *ServiceUnit) ListUnits(ctx context.Context) ([]domain.Unit, error) {
	return s.unit.ListUnits(ctx, s.storage)
}
