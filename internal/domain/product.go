package domain

import "context"

type ProductCreate struct {
	Title string
}

type ProductUpdate struct {
	Title *string
}

type ProductDetails struct {
	ID      int
	Title   string
	GroupID int
	Group   string
}

// -------------------------------------------------
// -------------------------------------------------

type ProductAliasUpdate struct {
	Alias *string
}

type ProductAliasDetails struct {
	ID        int
	ProductID int
	Product   string
	Alias     string
	GroupID   int
	Group     string
}

func (p ProductDetails) GetGroupID() int { return p.GroupID }
func (p ProductDetails) GetID() int      { return p.ID }

func (p ProductAliasDetails) GetGroupID() int { return p.GroupID }
func (p ProductAliasDetails) GetID() int      { return p.ID }

type ProductRepository interface {
	Create(ctx context.Context, q Querier, params any, groupID int) (ProductDetails, error)
	GetByID(ctx context.Context, q Querier, id int) (ProductDetails, error)
	UpdateByID(ctx context.Context, q Querier, id int, updates any) (ProductDetails, error)
	DeleteByID(ctx context.Context, q Querier, id int) error
	List(ctx context.Context, q Querier, groupID []int) ([]ProductDetails, error)
	ListAll(ctx context.Context, q Querier) ([]ProductDetails, error)
}

type ProductAliasRepository interface {
	Create(ctx context.Context, q Querier, productID int, alias string, groupID int) (ProductAliasDetails, error)
	GetByID(ctx context.Context, q Querier, aliasID int) (ProductAliasDetails, error)
	UpdateByID(ctx context.Context, q Querier, aliasID int, updateAlias ProductAliasUpdate) (ProductAliasDetails, error)
	DeleteByID(ctx context.Context, q Querier, aliasID int) error
	List(ctx context.Context, q Querier, productID int, groupID []int) ([]ProductAliasDetails, error)
	ListAll(ctx context.Context, q Querier, productID int) ([]ProductAliasDetails, error)
	DeleteAllProductAliases(ctx context.Context, q Querier, productID int) error
	FindProductByAlias(ctx context.Context, q Querier, alias string, groupID []int) (string, error)
	FindAllProductByAlias(ctx context.Context, q Querier, alias string) (string, error)
}
