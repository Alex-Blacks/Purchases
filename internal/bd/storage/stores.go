package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/jackc/pgx/v5"
)

type StoreRepo struct {
	storage *Storage
}

func NewStoreRepo(storage *Storage) *StoreRepo {
	return &StoreRepo{
		storage: storage,
	}
}

func (s *StoreRepo) CreateStore(ctx context.Context, name string) error {
	_, err := s.storage.pool.Exec(ctx, "INSERT INTO stores(name) VALUES($1)", name)
	if err != nil {
		return fmt.Errorf("Error created store: %w", err)
	}
	return nil
}

func (s *StoreRepo) GetStoreById(ctx context.Context, id int) (string, error) {
	var name string

	err := s.storage.pool.QueryRow(ctx, "SELECT name FROM stores WHERE id=$1", id).Scan(&name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrNotFound
		}
		return "", fmt.Errorf("Error get store: %w", err)
	}

	return name, nil
}

func (s *StoreRepo) DeleteStore(ctx context.Context, id int) error {
	tag, err := s.storage.pool.Exec(ctx, "DELETE FROM stores WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("Error delete store: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (s *StoreRepo) ListStore(ctx context.Context) ([]domain.StoreDTO, error) {
	rows, err := s.storage.pool.Query(ctx, "SELECT id,name FROM stores")
	if err != nil {
		return nil, fmt.Errorf("Error get list stores: %w", err)
	}
	defer rows.Close()

	var list []domain.StoreDTO

	for rows.Next() {
		var st domain.StoreDTO

		if err := rows.Scan(&st.Id, &st.Name); err != nil {
			return nil, fmt.Errorf("Error get list stores: %w", err)
		}

		list = append(list, st)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("Error get list stores: %w", rows.Err())
	}

	return list, nil
}
