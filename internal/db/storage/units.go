package storage

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

type UnitRepo struct{}

func NewUnitRepo() *UnitRepo {
	return &UnitRepo{}
}

func (u *UnitRepo) ListUnits(ctx context.Context, q domain.Querier) ([]domain.Unit, error) {
	rows, err := q.Query(ctx, `SELECT id,name,short_name FROM units`)
	if err != nil {
		return []domain.Unit{}, fmt.Errorf("query units: %w", err)
	}
	defer rows.Close()

	var units []domain.Unit
	for rows.Next() {
		var unit domain.Unit
		if err := rows.Scan(&unit.Id, &unit.Name, &unit.ShortName); err != nil {
			return []domain.Unit{}, fmt.Errorf("scan unit: %w", err)
		}

		units = append(units, unit)
	}
	if err := rows.Err(); err != nil {
		return []domain.Unit{}, fmt.Errorf("iteration failed: %w", err)
	}

	return units, nil
}
