package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type CategoryRepo struct{}

func NewCategoryRepo() *CategoryRepo {
	return &CategoryRepo{}
}

func (c *CategoryRepo) CreateCategory(ctx context.Context, q domain.Querier, name string) (domain.Category, error) {
	var category domain.Category
	if err := q.QueryRow(ctx, `INSERT INTO categories(name) VALUES ($1) RETURNING id`, name).Scan(&category.ID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.Category{}, domain.ErrAlreadyExists
		}
		return domain.Category{}, fmt.Errorf("create category: %w", err)
	}
	category.Name = name
	return category, nil
}

func (c *CategoryRepo) GetCategory(ctx context.Context, q domain.Querier, id int) (domain.Category, error) {
	var category domain.Category
	if err := q.QueryRow(ctx, `SELECT id, name FROM categories WHERE id = $1`, id).Scan(&category.ID, &category.Name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return category, domain.ErrNotFound
		}
		return category, fmt.Errorf("get category: %w", err)
	}
	return category, nil
}

func (c *CategoryRepo) DeleteCategory(ctx context.Context, q domain.Querier, id int) error {
	var categoryID int
	if err := q.QueryRow(ctx, `DELETE FROM categories WHERE id = $1 RETURNING id`, id).Scan(&categoryID); err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation:
			return domain.ErrConflict
		case errors.Is(err, pgx.ErrNoRows):
			return domain.ErrNotFound
		default:
			return fmt.Errorf("delete category: %w", err)
		}
	}
	return nil
}

func (c *CategoryRepo) ListCategories(ctx context.Context, q domain.Querier) ([]domain.Category, error) {
	rows, err := q.Query(ctx, `SELECT id, name FROM categories`)
	if err != nil {
		return nil, fmt.Errorf("query category: %w", err)
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var category domain.Category

		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}

		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration failed: %w", err)
	}

	return categories, nil
}
