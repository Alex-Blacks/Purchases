package domain

import "context"

type ProductAliasDetails struct {
	ID        int
	ProductID int
	Product   string
	Alias     string
	GroupID   int
	Group     string
}

type ProductAliasUpdate struct {
	ProductID *int
	Alias     *string
	GroupID   *int
}

type ProductDetails struct {
	ID      int
	Title   string
	GroupID int
	Group   string
}

type ProductUpdate struct {
	Title *string
}

func (p ProductDetails) GetGroupID() int      { return p.GroupID }
func (p ProductAliasDetails) GetGroupID() int { return p.GroupID }

type ProductRepository interface {
	CreateProduct(ctx context.Context, q Querier, title string, groupID int) (ProductDetails, error)
	GetProductByID(ctx context.Context, q Querier, productID int) (ProductDetails, error)
	UpdateProductByID(ctx context.Context, q Querier, productID int, updateProduct ProductUpdate) (ProductDetails, error)
	DeleteProductByID(ctx context.Context, q Querier, productID int) error
	ListProducts(ctx context.Context, q Querier, groupID []int) ([]ProductDetails, error)

	CreateProductAlias(ctx context.Context, q Querier, productID int, alias string, groupID int) (ProductAliasDetails, error)
	GetProductAliasByID(ctx context.Context, q Querier, aliasID int) (ProductAliasDetails, error)
	UpdateProductAliasByID(ctx context.Context, q Querier, aliasID int, updateAlias ProductAliasUpdate) (ProductAliasDetails, error)
	DeleteProductAliasByID(ctx context.Context, q Querier, aliasID int) error
	ListProductAliases(ctx context.Context, q Querier, productID int, groupID []int) ([]ProductAliasDetails, error)
	DeleteAllProductAliases(ctx context.Context, q Querier, productID int) error
	FindProductByAlias(ctx context.Context, q Querier, alias string) (string, error)
}
