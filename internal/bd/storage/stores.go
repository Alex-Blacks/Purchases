package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{
		pool: pool,
	}
}

type Store struct {
	id   int
	name string
}

func (s *Storage) CreateStore(ctx context.Context, name string) error {
	_, err := s.pool.Exec(ctx, "INSERT INTO stores(name) VALUES($1)", name)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) GetStoreById(ctx context.Context, id int) (string, error) {
	var name string

	err := s.pool.QueryRow(ctx, "SELECT name FROM stores WHERE id=$1", id).Scan(&name)
	if err != nil {
		return "", fmt.Errorf("Error get store by id: %w", err)
	}

	return name, nil
}

func (s *Storage) DeleteStore(ctx context.Context, id int) error {
	tag, err := s.pool.Exec(ctx, "DELETE FROM stores WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("Error delete store: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("store not found")
	}

	return nil
}

func (s *Storage) ListStores(ctx context.Context) ([]Store, error) {
	rows, err := s.pool.Query(ctx, "SELECT id,name FROM stores")
	if err != nil {
		return nil, fmt.Errorf("errors get list stores: %w", err)
	}
	defer rows.Close()

	var list []Store

	for rows.Next() {
		var st Store

		if err := rows.Scan(&st.id, &st.name); err != nil {
			return nil, fmt.Errorf("errors get list stores: %w", err)
		}

		list = append(list, st)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("errors get list stores: %w", rows.Err())
	}

	return list, nil
}
