package domain

import "context"

type ProductAliasDetails struct {
	ID      int
	Product string
	Alias   string
}

type ProductDetails struct {
	ID       int
	Title    string
	Unit     string
	Category string
}

type ProductRepository interface {
	CreateProduct(ctx context.Context, q Querier, title, unit string, categoryID int) (ProductDetails, error)
	GetProduct(ctx context.Context, q Querier, id int) (ProductDetails, error)
	DeleteProduct(ctx context.Context, q Querier, id int) error
	ListProducts(ctx context.Context, q Querier) ([]ProductDetails, error)

	CreateProductAlias(ctx context.Context, q Querier, productID int, alias string) (ProductAliasDetails, error)
	GetProductAlias(ctx context.Context, q Querier, id int) (ProductAliasDetails, error)
	DeleteProductAlias(ctx context.Context, q Querier, id int) error
	ListProductAliases(ctx context.Context, q Querier, productID int) ([]ProductAliasDetails, error)
	DeleteAllProductAliases(ctx context.Context, q Querier, productID int) error
	FindProductByAlias(ctx context.Context, q Querier, alias string) (string, error)
}
