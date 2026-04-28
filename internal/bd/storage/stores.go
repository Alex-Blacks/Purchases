package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (s *Storage) CreateStore(ctx context.Context, name string) error {
	_, err := s.pool.Exec(ctx, "INSERT INTO stores(name) VALUES($1)", name)
	if err != nil {
		return fmt.Errorf("Error created store: %w", err)
	}
	return nil
}

func (s *Storage) GetStoreById(ctx context.Context, id int) (string, error) {
	var name string

	err := s.pool.QueryRow(ctx, "SELECT name FROM stores WHERE id=$1", id).Scan(&name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrNotFound
		}
		return "", fmt.Errorf("Error get store: %w", err)
	}

	return name, nil
}

func (s *Storage) DeleteStore(ctx context.Context, id int) error {
	tag, err := s.pool.Exec(ctx, "DELETE FROM stores WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("Error delete store: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (s *Storage) ListStore(ctx context.Context) ([]domain.ListStore, error) {
	rows, err := s.pool.Query(ctx, "SELECT id,name FROM stores")
	if err != nil {
		return nil, fmt.Errorf("errors get list stores: %w", err)
	}
	defer rows.Close()

	var list []domain.ListStore

	for rows.Next() {
		var st domain.ListStore

		if err := rows.Scan(&st.Id, &st.Name); err != nil {
			return nil, fmt.Errorf("errors get list stores: %w", err)
		}

		list = append(list, st)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("errors get list stores: %w", rows.Err())
	}

	return list, nil
}
