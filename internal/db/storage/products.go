package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ProductRepo struct{}

func NewProductRepo() *ProductRepo {
	return &ProductRepo{}
}

func (p *ProductRepo) CreateProduct(ctx context.Context, q domain.Querier, title string, groupID int) (domain.ProductDetails, error) {
	var product domain.ProductDetails
	if err := q.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO products(title, group_id) 
			VALUES ($1,$2) 
			RETURNING id, title, group_id
		)
		SELECT i.id, i.title, i.group_id, g.name
		FROM inserted i
		JOIN groups g ON i.group_id = g.id
	`, title, groupID).Scan(&product.ID, &product.Title, &product.GroupID, &product.Group); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolation:
				return domain.ProductDetails{}, domain.ErrAlreadyExists
			case pgForeignKeyViolation:
				return domain.ProductDetails{}, domain.ErrConflict
			}
		}
		return domain.ProductDetails{}, fmt.Errorf("create product: %w", err)
	}
	return product, nil
}

func (p *ProductRepo) GetProductByID(ctx context.Context, q domain.Querier, productID int) (domain.ProductDetails, error) {
	var product domain.ProductDetails
	if err := q.QueryRow(ctx, `
		SELECT p.id, p.title, p.group_id, g.name
		FROM products p
		JOIN groups g ON p.group_id = g.id
		WHERE p.id = $1	
	`, productID).Scan(&product.ID, &product.Title, &product.GroupID, &product.Group); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return product, domain.ErrNotFound
		}
		return product, fmt.Errorf("get product: %w", err)
	}
	return product, nil
}

func (p *ProductRepo) UpdateProductByID(ctx context.Context, q domain.Querier, productID int, updateProduct domain.ProductUpdate) (domain.ProductDetails, error) {
	var product domain.ProductDetails
	args := []any{productID}
	setParts := []string{}
	argPos := 2

	if (updateProduct.Title != nil) && (strings.TrimSpace(*updateProduct.Title) != "") {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argPos))
		args = append(args, *updateProduct.Title)
		argPos++
	}

	set := strings.Join(setParts, ", ")
	if strings.TrimSpace(set) == "" {
		return domain.ProductDetails{}, domain.ErrNoFieldsToUpdate
	}

	if err := q.QueryRow(ctx, `
		UPDATE products p
		SET `+set+`
		FROM groups g
		WHERE p.id = $1 AND p.group_id = g.id
		RETURNING p.id, p.title, p.group_id, g.name
	`, args...).Scan(&product.ID, &product.Title, &product.GroupID, &product.Group); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProductDetails{}, domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolation:
				return domain.ProductDetails{}, domain.ErrAlreadyExists
			case pgForeignKeyViolation:
				return domain.ProductDetails{}, domain.ErrConflict
			}
		}
		return domain.ProductDetails{}, fmt.Errorf("update product: %w", err)
	}
	return product, nil
}

func (p *ProductRepo) DeleteProductByID(ctx context.Context, q domain.Querier, productID int) error {
	var id int
	if err := q.QueryRow(ctx, `DELETE FROM products WHERE products.id = $1 RETURNING id`, productID).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
			return domain.ErrConflict
		}
		return fmt.Errorf("delete product: %w", err)
	}
	return nil
}

func (p *ProductRepo) ListProducts(ctx context.Context, q domain.Querier, groupID []int) ([]domain.ProductDetails, error) {
	rows, err := q.Query(ctx, `
		SELECT p.id, p.title, p.group_id, g.name
		FROM products p
		JOIN groups g ON p.group_id = g.id
		WHERE p.group_id = ANY($1::int[])
	`)
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	var products []domain.ProductDetails
	for rows.Next() {
		var product domain.ProductDetails
		if err := rows.Scan(&product.ID, &product.Title, &product.GroupID, &product.Group); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration failed: %w", err)
	}

	return products, nil
}

func (a *ProductRepo) CreateProductAlias(ctx context.Context, q domain.Querier, productID int, alias string, groupID int) (domain.ProductAliasDetails, error) {
	var productAlias domain.ProductAliasDetails
	if err := q.QueryRow(ctx, `
		WITH inserted AS (	
			INSERT INTO product_aliases(product_id, alias, group_id) 
			VALUES ($1,$2,$3) 
			RETURNING id, product_id, alias, group_id
		)
		SELECT i.id, i.product_id, p.title, i.alias, i.group_id, g.name
		FROM inserted i
		JOIN products p ON i.product_id = p.id
		JOIN groups g ON i.group_id = g.id
	`, productID, alias, groupID).Scan(&productAlias.ID, &productAlias.ProductID, &productAlias.Product, &productAlias.Alias, &productAlias.GroupID, &productAlias.Group); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolation:
				return domain.ProductAliasDetails{}, domain.ErrAlreadyExists
			case pgForeignKeyViolation:
				return domain.ProductAliasDetails{}, domain.ErrConflict
			}
		}
		return domain.ProductAliasDetails{}, fmt.Errorf("create product alias: %w", err)
	}

	return productAlias, nil
}
func (a *ProductRepo) GetProductAliasByID(ctx context.Context, q domain.Querier, aliasID int) (domain.ProductAliasDetails, error) {
	var alias domain.ProductAliasDetails
	if err := q.QueryRow(ctx, `
		SELECT pa.id, pa.product_id, p.title, pa.alias, pa.group_id, g.name
		FROM product_aliases pa
		JOIN products p ON pa.product_id = p.id
		JOIN groups g ON pa.group_id = g.id
		WHERE pa.id = $1
	`, aliasID).Scan(&alias.ID, &alias.ProductID, &alias.Product, &alias.Alias, &alias.GroupID, &alias.Group); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return alias, domain.ErrNotFound
		}
		return alias, fmt.Errorf("get product alias: %w", err)
	}
	return alias, nil
}

func (a *ProductRepo) UpdateProductAliasByID(ctx context.Context, q domain.Querier, aliasID int, updateAlias domain.ProductAliasUpdate) (domain.ProductAliasDetails, error) {
	var alias domain.ProductAliasDetails
	args := []any{aliasID}
	setParts := []string{}
	argPos := 2

	if updateAlias.ProductID != nil && *updateAlias.ProductID >= 1 {
		setParts = append(setParts, fmt.Sprintf("product_id = $%d", argPos))
		args = append(args, *updateAlias.ProductID)
		argPos++
	}
	if updateAlias.Alias != nil && strings.TrimSpace(*updateAlias.Alias) != "" {
		setParts = append(setParts, fmt.Sprintf("alias = $%d", argPos))
		args = append(args, *updateAlias.Alias)
		argPos++
	}
	if updateAlias.GroupID != nil && *updateAlias.GroupID >= 1 {
		setParts = append(setParts, fmt.Sprintf("group_id = $%d", argPos))
		args = append(args, *updateAlias.GroupID)
		argPos++
	}

	set := strings.Join(setParts, ", ")
	if strings.TrimSpace(set) == "" {
		return domain.ProductAliasDetails{}, domain.ErrNoFieldsToUpdate
	}

	if err := q.QueryRow(ctx, `
		UPDATE product_aliases pa
		SET `+set+`
		FROM products p
		JOIN groups g ON pa.group_id = g.id
		WHERE pa.id = $1 AND pa.product_id = p.id
		RETURNING pa.id, pa.product_id, p.title, pa.alias, pa.group_id, g.name
	`, args...).Scan(&alias.ID, &alias.ProductID, &alias.Product, &alias.Alias, &alias.GroupID, &alias.Group); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProductAliasDetails{}, domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolation:
				return domain.ProductAliasDetails{}, domain.ErrAlreadyExists
			case pgForeignKeyViolation:
				return domain.ProductAliasDetails{}, domain.ErrConflict
			}
		}
		return domain.ProductAliasDetails{}, fmt.Errorf("update product alias: %w", err)
	}
	return alias, nil
}

func (a *ProductRepo) DeleteProductAliasByID(ctx context.Context, q domain.Querier, aliasID int) error {
	var id int
	if err := q.QueryRow(ctx, `DELETE FROM product_aliases WHERE id = $1 RETURNING id`, aliasID).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
			return domain.ErrConflict
		}
		return fmt.Errorf("delete product alias: %w", err)
	}
	return nil
}
func (a *ProductRepo) ListProductAliases(ctx context.Context, q domain.Querier, productID int, groupID []int) ([]domain.ProductAliasDetails, error) {
	rows, err := q.Query(ctx, `
		SELECT pa.id, p.title, pa.alias, pa.group_id, g.name
		FROM product_aliases pa
		JOIN products p ON pa.product_id = p.id
		JOIN groups g ON pa.group_id = g.id
		WHERE pa.product_id = $1 AND pa.group_id = ANY($1::int[])
	`, productID)
	if err != nil {
		return nil, fmt.Errorf("query product aliases: %w", err)
	}
	defer rows.Close()

	var aliases []domain.ProductAliasDetails
	for rows.Next() {
		var alias domain.ProductAliasDetails
		if err := rows.Scan(&alias.ID, &alias.ProductID, &alias.Product, &alias.Alias, &alias.GroupID, &alias.Group); err != nil {
			return nil, fmt.Errorf("product aliases: %w", err)
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolation:
				return domain.ErrAlreadyExists
			case pgForeignKeyViolation:
				return domain.ErrConflict
			}
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
