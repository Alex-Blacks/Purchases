package storage

import (
	"context"
	"fmt"
)

func (s *Storage) CreateCategory(ctx context.Context, name string) error {
	_, err := s.pool.Exec(ctx, "INSERT INTO categories(name) VALUES($1)", name)
	if err != nil {
		return fmt.Errorf("Error create category: %w", err)
	}

	return nil
}

func (s *Storage) GetCategoryById(ctx context.Context, id int) (string, error) {
	var name string
	if err := s.pool.QueryRow(ctx, "SELECT name FROM category WHERE id = $1", id).Scan(&name); err != nil {
		return "", fmt.Errorf("Error get category by id: %w", err)
	}

	return name, nil
}

func (s *Storage) DeleteCategory(ctx context.Context, id int) error {
	tag, err := s.pool.Exec()
}
