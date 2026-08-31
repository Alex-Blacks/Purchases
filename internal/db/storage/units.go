package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UnitRepo struct{}

func NewUnitRepo() *UnitRepo {
	return &UnitRepo{}
}

func (u *UnitRepo) Create(ctx context.Context, q domain.Querier, params any, groupID int) (domain.UnitDetails, error) {
	unitCreate, ok := params.(domain.UnitCreate)
	if !ok {
		return domain.UnitDetails{}, fmt.Errorf("invalid params type: expected UnitCreate, got %T", params)
	}
	var unit domain.UnitDetails
	if err := q.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO units(name, short_name, group_id) 
			VALUES ($1,$2,$3) 
			RETURNING id, name, short_name, group_id
		)
		SELECT i.id, i.name, i.short_name, i.group_id, g.name
		FROM inserted i
		JOIN groups g ON i.group_id = g.id
	`, unitCreate.Name, unitCreate.ShortName, groupID).Scan(&unit.ID, &unit.Name, &unit.ShortName, &unit.GroupID, &unit.Group); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolation:
				return domain.UnitDetails{}, domain.ErrAlreadyExists
			case pgForeignKeyViolation:
				return domain.UnitDetails{}, domain.ErrConflict
			}
		}
		return domain.UnitDetails{}, fmt.Errorf("create unit: %w", err)
	}
	return unit, nil
}

func (u *UnitRepo) GetByID(ctx context.Context, q domain.Querier, id int) (domain.UnitDetails, error) {
	var unit domain.UnitDetails
	if err := q.QueryRow(ctx, `
		SELECT u.id, u.name, u.short_name, u.group_id, g.name
		FROM units u
		JOIN groups g ON u.group_id = g.id
		WHERE u.id = $1
	`, id).Scan(&unit.ID, &unit.Name, &unit.ShortName, &unit.GroupID, &unit.Group); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.UnitDetails{}, domain.ErrNotFound
		}
		return domain.UnitDetails{}, fmt.Errorf("get unit: %w", err)
	}
	return unit, nil
}

func (u *UnitRepo) UpdateByID(ctx context.Context, q domain.Querier, id int, updates any) (domain.UnitDetails, error) {
	unitUpdate, ok := updates.(domain.UnitUpdate)
	if !ok {
		return domain.UnitDetails{}, fmt.Errorf("invalid updates type: expected UnitUpdate, got %T", updates)
	}

	args := []any{id}
	setParts := []string{}
	argPos := 2

	if unitUpdate.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argPos))
		args = append(args, *unitUpdate.Name)
		argPos++
	}
	if unitUpdate.ShortName != nil {
		setParts = append(setParts, fmt.Sprintf("short_name = $%d", argPos))
		args = append(args, *unitUpdate.ShortName)
		argPos++
	}

	set := strings.Join(setParts, ", ")
	if strings.TrimSpace(set) == "" {
		return domain.UnitDetails{}, domain.ErrNoFieldsToUpdate
	}

	var unit domain.UnitDetails
	if err := q.QueryRow(ctx, `
		UPDATE units u
		SET `+set+`
		FROM groups g
		WHERE u.id = $1 AND u.group_id = g.id
		RETURNING u.id, u.name, u.short_name, u.group_id, g.name
	`, args...).Scan(&unit.ID, &unit.Name, &unit.ShortName, &unit.GroupID, &unit.Group); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.UnitDetails{}, domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolation:
				return domain.UnitDetails{}, domain.ErrAlreadyExists
			case pgForeignKeyViolation:
				return domain.UnitDetails{}, domain.ErrConflict
			}
		}
		return domain.UnitDetails{}, fmt.Errorf("update unit: %w", err)
	}
	return unit, nil
}

func (u *UnitRepo) DeleteByID(ctx context.Context, q domain.Querier, id int) error {
	var deleteID int
	if err := q.QueryRow(ctx, `DELETE FROM units WHERE units.id = $1 RETURNING id`, id).Scan(&deleteID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
			return domain.ErrConflict
		}
		return fmt.Errorf("delete unit: %w", err)
	}
	return nil
}

func (u *UnitRepo) List(ctx context.Context, q domain.Querier, filter domain.UnitListFilter) ([]domain.UnitDetails, error) {
	query := `
		SELECT u.id, u.name, u.short_name, u.group_id, g.name 
		FROM units u
		JOIN groups g ON u.group_id = g.id`

	whereClause, whereArgs, whereArgPos := buildUnitWhere(filter)
	query += whereClause

	query += fmt.Sprintf("ORDER BY u.id LIMIT $%d OFFSET $%d", whereArgPos, whereArgPos+1)
	args := append(whereArgs, filter.Limit, filter.Offset)

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query units: %w", err)
	}
	defer rows.Close()

	var units []domain.UnitDetails
	for rows.Next() {
		var unit domain.UnitDetails
		if err := rows.Scan(&unit.ID, &unit.Name, &unit.ShortName, &unit.GroupID, &unit.Group); err != nil {
			return nil, fmt.Errorf("scan unit: %w", err)
		}

		units = append(units, unit)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration failed: %w", err)
	}

	return units, nil
}

func (u *UnitRepo) Count(ctx context.Context, q domain.Querier, filter domain.UnitListFilter) (int, error) {
	query := "SELECT COUNT(*) FROM units u"

	whereClause, args, _ := buildUnitWhere(filter)
	query += whereClause

	var count int
	if err := q.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("query count units: %w", err)
	}

	return count, nil
}
