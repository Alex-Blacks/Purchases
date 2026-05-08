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
	if err := q.QueryRow(ctx, "INSERT INTO stores(name) VALUES($1) RETURNING id", name).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return 0, domain.ErrAlreadyExists
		}
		return 0, fmt.Errorf("created store: %w", err)
	}
	return id, nil
}

func (s *StoreRepo) GetStore(ctx context.Context, q domain.Querier, id int) (domain.Store, error) {
	var store domain.Store
	if err := q.QueryRow(ctx, "SELECT id,name FROM stores WHERE id=$1", id).Scan(&store.ID, &store.Name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store, domain.ErrNotFound
		}
		return store, fmt.Errorf("get store: %w", err)
	}
	return store, nil
}

func (s *StoreRepo) DeleteStore(ctx context.Context, q domain.Querier, id int) error {
	tag, err := q.Exec(ctx, "DELETE FROM stores WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete store: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (s *StoreRepo) ListStores(ctx context.Context, q domain.Querier) ([]domain.Store, error) {
	rows, err := q.Query(ctx, "SELECT id,name FROM stores")
	if err != nil {
		return nil, fmt.Errorf("get list stores: %w", err)
	}
	defer rows.Close()

	var list []domain.Store
	for rows.Next() {
		var st domain.Store

		if err := rows.Scan(&st.ID, &st.Name); err != nil {
			return nil, fmt.Errorf("get list stores: %w", err)
		}

		list = append(list, st)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("get list stores: %w", rows.Err())
	}

	return list, nil
}
