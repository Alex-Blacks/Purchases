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

func (p *ProductRepo) CreateProduct(ctx context.Context, q domain.Querier, title, unit string, categoryID int) (domain.ProductDetails, error) {
	var product domain.ProductDetails
	if err := q.QueryRow(ctx, `INSERT INTO products(title,unit,category_id) VALUES ($1,$2,$3) RETURNING id`, title, unit, categoryID).Scan(&product.ID); err != nil {
		var pgErr *pgconn.PgError
		switch errors.As(err, &pgErr) {
		case pgErr.Code == pgUniqueViolation:
			return domain.ProductDetails{}, domain.ErrAlreadyExists
		case pgErr.Code == pgForeignKeyViolation:
			return domain.ProductDetails{}, domain.ErrConflict
		default:
			return domain.ProductDetails{}, fmt.Errorf("query create product: %w", err)
		}
	}
	productOutput, err := p.GetProduct(ctx, q, product.ID)
	if err != nil {
		return domain.ProductDetails{}, fmt.Errorf("get product: %w", err)
	}
	return productOutput, nil
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

func (a *ProductRepo) CreateProductAlias(ctx context.Context, q domain.Querier, productID int, alias string) (domain.ProductAliasDetails, error) {
	var productAlias domain.ProductAliasDetails
	if err := q.QueryRow(ctx, `INSERT INTO product_aliases(product_id,alias) VALUES ($1,$2) RETURNING id`, productID, alias).Scan(&productAlias.ID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.ProductAliasDetails{}, domain.ErrAlreadyExists
		}
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
			return domain.ProductAliasDetails{}, domain.ErrConflict
		}
		return domain.ProductAliasDetails{}, fmt.Errorf("query create product alias: %w", err)
	}

	aliasOutput, err := a.GetProductAlias(ctx, q, productAlias.ID)
	if err != nil {
		return domain.ProductAliasDetails{}, fmt.Errorf("get product alias: %w", err)
	}
	return aliasOutput, nil
}
func (a *ProductRepo) GetProductAlias(ctx context.Context, q domain.Querier, id int) (domain.ProductAliasDetails, error) {
	var alias domain.ProductAliasDetails
	if err := q.QueryRow(ctx, `
		SELECT pa.id, p.title, pa.alias
		FROM product_aliases pa
		JOIN products p ON pa.product_id = p.id
		WHERE pa.id = $1
	`, id).Scan(&alias.ID, &alias.Product, &alias.Alias); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return alias, domain.ErrNotFound
		}
		return alias, fmt.Errorf("query scan product alias: %w", err)
	}
	return alias, nil
}
func (a *ProductRepo) DeleteProductAlias(ctx context.Context, q domain.Querier, id int) error {
	var aliasID int
	if err := q.QueryRow(ctx, `DELETE FROM product_aliases WHERE id = $1 RETURNING id`, id).Scan(&aliasID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("delete product alias: %w", err)
	}
	return nil
}
func (a *ProductRepo) ListProductAliases(ctx context.Context, q domain.Querier, productID int) ([]domain.ProductAliasDetails, error) {
	rows, err := q.Query(ctx, `
		SELECT pa.id, p.title, pa.alias
		FROM product_aliases pa
		JOIN products p ON pa.product_id = p.id
		WHERE pa.product_id = $1
	`, productID)
	if err != nil {
		return nil, fmt.Errorf("query product aliases: %w", err)
	}
	defer rows.Close()

	var aliases []domain.ProductAliasDetails
	for rows.Next() {
		var alias domain.ProductAliasDetails

		if err := rows.Scan(&alias.ID, &alias.Product, &alias.Alias); err != nil {
			return nil, fmt.Errorf("scan product aliases: %w", err)
		}

		aliases = append(aliases, alias)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration failed: %w", err)
	}

	return aliases, nil

}
func (a *ProductRepo) DeleteAllProductAliases(ctx context.Context, q domain.Querier, productID int) error {
	tag, err := q.Exec(ctx, `DELETE FROM product_aliases WHERE product_id = $1`, productID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("delete product alias: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (a *ProductRepo) FindProductByAlias(ctx context.Context, q domain.Querier, alias string) (string, error) {
	var product string
	if err := q.QueryRow(ctx, `
		SELECT p.title
		FROM product_aliases pa
		JOIN products p ON pa.product_id = p.id
		WHERE pa.alias = $1
	`, alias).Scan(&product); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrNotFound
		}
		return "", fmt.Errorf("query product alias: %w", err)
	}

	return product, nil
}
