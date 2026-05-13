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

func (s *StoreRepo) CreateStore(ctx context.Context, q domain.Querier, name string) (int, error) {
	var id int
	if err := q.QueryRow(ctx, `INSERT INTO stores(name) VALUES($1) RETURNING id`, name).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return 0, domain.ErrAlreadyExists
		}
		return 0, fmt.Errorf("create store: %w", err)
	}
	return id, nil
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
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("delete store: %w", err)
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
