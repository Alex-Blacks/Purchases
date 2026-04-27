package storage

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
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
	tag, err := s.pool.Exec(ctx, "DELETE FROM categories WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("Error delete category: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

func (s *Storage) ListCategories(ctx context.Context) ([]domain.ListStore, error) {
	rows, err := s.pool.Query(ctx, "SELECT id, name FROM categories")
	if err != nil {
		return nil, fmt.Errorf("Error get category: %w", err)
	}
	defer rows.Close()

	var list []domain.ListStore

	for rows.Next() {
		var st domain.ListStore

		if err := rows.Scan(&st.Id, &st.Name); err != nil {
			return nil, fmt.Errorf("Error get category: %w", err)
		}

		list = append(list, st)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error get category: %w", err)
	}

	return list, nil
}
