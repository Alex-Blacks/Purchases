package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type StoreRepo struct{}

func NewStoreRepo() *StoreRepo {
	return &StoreRepo{}
}

func (s *StoreRepo) CreateStore(ctx context.Context, q domain.Querier, name string) (domain.Store, error) {
	var store domain.Store
	if err := q.QueryRow(ctx, `INSERT INTO stores(name) VALUES($1) RETURNING id`, name).Scan(&store.ID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.Store{}, domain.ErrAlreadyExists
		}
		return domain.Store{}, fmt.Errorf("create store: %w", err)
	}
	store.Name = name
	return store, nil
}

func (s *StoreRepo) GetStore(ctx context.Context, q domain.Querier, id int) (domain.Store, error) {
	var store domain.Store
	if err := q.QueryRow(ctx, `SELECT id,name FROM stores WHERE id=$1`, id).Scan(&store.ID, &store.Name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store, domain.ErrNotFound
		}
		return store, fmt.Errorf("get store: %w", err)
	}
	return store, nil
}

func (s *StoreRepo) DeleteStore(ctx context.Context, q domain.Querier, id int) error {
	var storeID int
	if err := q.QueryRow(ctx, `DELETE FROM stores WHERE id = $1 RETURNING id`, id).Scan(&storeID); err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation:
			return domain.ErrConflict
		case errors.Is(err, pgx.ErrNoRows):
			return domain.ErrNotFound
		default:
			return fmt.Errorf("delete store: %w", err)
		}
	}

	return nil
}

func (s *StoreRepo) ListStores(ctx context.Context, q domain.Querier) ([]domain.Store, error) {
	rows, err := q.Query(ctx, `SELECT id,name FROM stores`)
	if err != nil {
		return nil, fmt.Errorf("query stores: %w", err)
	}
	defer rows.Close()

	var stores []domain.Store
	for rows.Next() {
		var store domain.Store

		if err := rows.Scan(&store.ID, &store.Name); err != nil {
			return nil, fmt.Errorf("scan store: %w", err)
		}

		stores = append(stores, store)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration failed: %w", rows.Err())
	}

	return stores, nil
}
