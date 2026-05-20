package domain

import "context"

type ProductAliasDetails struct {
	ID      int
	Product string
	Alias   string
}

type ProductAliasRepository interface {
	CreateProductAlias(ctx context.Context, q Querier, productID int, alias string) (int, error)
	GetProductAlias(ctx context.Context, q Querier, id int) (ProductAliasDetails, error)
	DeleteProductAlias(ctx context.Context, q Querier, id int) error
	ListProductAliases(ctx context.Context, q Querier, productID int) ([]ProductAliasDetails, error)
	DeleteAllProductAliases(ctx context.Context, q Querier, productID int) error
	FindProductByAlias(ctx context.Context, q Querier, alias string) (string, error)
}

// ---------------------------------------------------------------------------------------------------------------------------------

type ProductDetails struct {
	ID       int
	Title    string
	Unit     string
	Category string
}

type ProductRepository interface {
	CreateProduct(ctx context.Context, q Querier, title, unit string, categoryID int) (int, error)
	GetProduct(ctx context.Context, q Querier, id int) (ProductDetails, error)
	DeleteProduct(ctx context.Context, q Querier, id int) error
	ListProducts(ctx context.Context, q Querier) ([]ProductDetails, error)
}
