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

type StoreRepo struct{}

func NewStoreRepo() *StoreRepo {
	return &StoreRepo{}
}

func (s *StoreRepo) Create(ctx context.Context, q domain.Querier, params any, groupID int) (domain.StoreDetails, error) {
	storeCreate, ok := params.(domain.StoreCreate)
	if !ok {
		return domain.StoreDetails{}, fmt.Errorf("invalid params type: expected StoreCreate, got %T", params)
	}
	var store domain.StoreDetails
	if err := q.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO stores(name, group_id) 
			VALUES($1,$2) 
			RETURNING id, name, group_id
		)
		SELECT i.id, i.name, i.group_id, g.name
		FROM inserted i
		JOIN groups g ON i.group_id = g.id
	`, storeCreate.Name, groupID).Scan(&store.ID, &store.Name, &store.GroupID, &store.Group); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolation:
				return domain.StoreDetails{}, domain.ErrAlreadyExists
			case pgForeignKeyViolation:
				return domain.StoreDetails{}, domain.ErrConflict
			}
		}
		return domain.StoreDetails{}, fmt.Errorf("create store: %w", err)
	}

	return store, nil
}

func (s *StoreRepo) GetByID(ctx context.Context, q domain.Querier, id int) (domain.StoreDetails, error) {
	var store domain.StoreDetails
	if err := q.QueryRow(ctx, `
		SELECT s.id, s.name, s.group_id, g.name 
		FROM stores s
		JOIN groups g ON s.group_id = g.id
		WHERE s.id=$1
	`, id).Scan(&store.ID, &store.Name, &store.GroupID, &store.Group); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.StoreDetails{}, domain.ErrNotFound
		}
		return domain.StoreDetails{}, fmt.Errorf("get store: %w", err)
	}
	return store, nil
}

func (s *StoreRepo) UpdateByID(ctx context.Context, q domain.Querier, id int, updates any) (domain.StoreDetails, error) {
	storeUpdate, ok := updates.(domain.StoreUpdate)
	if !ok {
		return domain.StoreDetails{}, fmt.Errorf("invalid updates type: expected StoreUpdate, got %T", updates)
	}
	var store domain.StoreDetails
	args := []any{id}
	setParts := []string{}
	argPos := 2

	if (storeUpdate.Name != nil) && (strings.TrimSpace(*storeUpdate.Name) != "") {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argPos))
		args = append(args, *storeUpdate.Name)
		argPos++
	}

	set := strings.Join(setParts, ", ")
	if strings.TrimSpace(set) == "" {
		return domain.StoreDetails{}, domain.ErrNoFieldsToUpdate
	}

	if err := q.QueryRow(ctx, `
		UPDATE stores s
		SET `+set+`
		FROM groups g
		WHERE s.id = $1 AND s.group_id = g.id
		RETURNING s.id, s.name, s.group_id, g.name
	`, args...).Scan(&store.ID, &store.Name, &store.GroupID, &store.Group); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.StoreDetails{}, domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolation:
				return domain.StoreDetails{}, domain.ErrAlreadyExists
			case pgForeignKeyViolation:
				return domain.StoreDetails{}, domain.ErrConflict
			}
		}
		return domain.StoreDetails{}, fmt.Errorf("update store: %w", err)
	}
	return store, nil
}

func (s *StoreRepo) DeleteByID(ctx context.Context, q domain.Querier, id int) error {
	var deleteID int
	if err := q.QueryRow(ctx, `DELETE FROM stores WHERE stores.id = $1 RETURNING id`, id).Scan(&deleteID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
			return domain.ErrConflict
		}
		return fmt.Errorf("delete store: %w", err)
	}

	return nil
}

func (s *StoreRepo) List(ctx context.Context, q domain.Querier, groupID []int) ([]domain.StoreDetails, error) {
	rows, err := q.Query(ctx, `
		SELECT s.id, s.name, s.group_id, g.name 
		FROM stores s
		JOIN groups g ON s.group_id = g.id
		WHERE s.group_id = ANY($1::int[])
	`, groupID)
	if err != nil {
		return []domain.StoreDetails{}, fmt.Errorf("query stores: %w", err)
	}
	defer rows.Close()

	var stores []domain.StoreDetails
	for rows.Next() {
		var store domain.StoreDetails

		if err := rows.Scan(&store.ID, &store.Name, &store.GroupID, &store.Group); err != nil {
			return []domain.StoreDetails{}, fmt.Errorf("scan store: %w", err)
		}

		stores = append(stores, store)
	}

	if err = rows.Err(); err != nil {
		return []domain.StoreDetails{}, fmt.Errorf("iteration failed: %w", rows.Err())
	}

	return stores, nil
}

func (s *StoreRepo) ListAll(ctx context.Context, q domain.Querier) ([]domain.StoreDetails, error) {
	rows, err := q.Query(ctx, `
		SELECT s.id, s.name, s.group_id, g.name 
		FROM stores s
		JOIN groups g ON s.group_id = g.id
	`)
	if err != nil {
		return []domain.StoreDetails{}, fmt.Errorf("query stores: %w", err)
	}
	defer rows.Close()

	var stores []domain.StoreDetails
	for rows.Next() {
		var store domain.StoreDetails

		if err := rows.Scan(&store.ID, &store.Name, &store.GroupID, &store.Group); err != nil {
			return []domain.StoreDetails{}, fmt.Errorf("scan store: %w", err)
		}

		stores = append(stores, store)
	}

	if err = rows.Err(); err != nil {
		return []domain.StoreDetails{}, fmt.Errorf("iteration failed: %w", rows.Err())
	}

	return stores, nil
}
