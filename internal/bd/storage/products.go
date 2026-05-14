package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ProductRepo struct{}

func NewProductRepo() *ProductRepo {
	return &ProductRepo{}
}

func (p *ProductRepo) CreateProduct(ctx context.Context, q domain.Querier, title, unit string, categoryID int) (int, error) {
	var id int
	if err := q.QueryRow(ctx, `INSERT INTO products(title,unit,category_id) VALUES ($1,$2,$3) RETURNING id`, title, unit, categoryID).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return 0, domain.ErrAlreadyExists
		}
		return 0, fmt.Errorf("query create product: %w", err)
	}

	return id, nil
}

func (p *ProductRepo) GetProduct(ctx context.Context, q domain.Querier, id int) (domain.ProductDetails, error) {
	var product domain.ProductDetails
	if err := q.QueryRow(ctx, `
		SELECT p.id, p.title, p.unit, c.name
		FROM products p
		JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1	
	`, id).Scan(&product.ID, &product.Title, &product.Unit, &product.Category); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return product, domain.ErrNotFound
		}
		return product, fmt.Errorf("get product: %w", err)
	}
	return product, nil
}

func (p *ProductRepo) DeleteProduct(ctx context.Context, q domain.Querier, id int) error {
	var productID int
	if err := q.QueryRow(ctx, `DELETE FROM products WHERE id = $1 RETURNING id`, id).Scan(&productID); err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation:
			return domain.ErrConflict
		case errors.Is(err, pgx.ErrNoRows):
			return domain.ErrNotFound
		default:
			return fmt.Errorf("delete product: %w", err)
		}
	}
	return nil
}

func (p *ProductRepo) ListProducts(ctx context.Context, q domain.Querier) ([]domain.ProductDetails, error) {
	rows, err := q.Query(ctx, `
		SELECT p.id, p.title, p.unit, c.name
		FROM products p
		JOIN categories c ON p.category_id = c.id
	`)
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	var products []domain.ProductDetails
	for rows.Next() {
		var product domain.ProductDetails

		if err := rows.Scan(&product.ID, &product.Title, &product.Unit, &product.Category); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration failed: %w", err)
	}

	return products, nil
}
