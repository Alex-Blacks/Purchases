package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/jackc/pgx/v5"
)

type ProductAliasRepo struct{}

func NewProductAliasRepo() *ProductAliasRepo {
	return &ProductAliasRepo{}
}

func (a *ProductAliasRepo) CreateProductAlias(ctx context.Context, q domain.Querier, productID int, alias string) (int, error) {
	var id int
	if err := q.QueryRow(ctx, `INSERT INTO product_aliases(product_id,alias) VALUES ($1,$2) RETURNING id`, productID, alias).Scan(&id); err != nil {
		return 0, fmt.Errorf("query create product alias: %w", err)
	}
	return id, nil
}
func (a *ProductAliasRepo) GetProductAlias(ctx context.Context, q domain.Querier, id int) (domain.ProductAliasDetails, error) {
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
func (a *ProductAliasRepo) DeleteProductAlias(ctx context.Context, q domain.Querier, id int) error {
	var aliasID int
	if err := q.QueryRow(ctx, `DELETE FROM product_aliases WHERE id = $1 RETURNING id`, id).Scan(&aliasID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("delete product alias: %w", err)
	}
	return nil
}
func (a *ProductAliasRepo) ListProductAliases(ctx context.Context, q domain.Querier, productID int) ([]domain.ProductAliasDetails, error) {
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
func (a *ProductAliasRepo) DeleteAllProductAliases(ctx context.Context, q domain.Querier, productID int) error {
	tag, err := q.Exec(ctx, `DELETE FROM product_aliases WHERE product_id = $1 RETURNING id`, productID)
	if err != nil {
		return fmt.Errorf("delete product alias: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
