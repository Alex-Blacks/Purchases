package dto

import (
	"time"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

//-------------------------------------------------------------------------------------------------

type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

//-------------------------------------------------------------------------------------------------

type UserRequest struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Role     string `json:"role,omitempty"`
}

type UserResponse struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

func ToUserResponse(user domain.User) UserResponse {
	return UserResponse{
		ID:     user.ID,
		Name:   user.Name,
		Email:  user.Email,
		Role:   user.Role,
		Status: user.Status,
	}
}

func ToUsersResponse(user []domain.User) []UserResponse {
	resp := make([]UserResponse, len(user))

	for i, it := range user {
		resp[i] = UserResponse{
			ID:     it.ID,
			Name:   it.Name,
			Email:  it.Email,
			Role:   it.Role,
			Status: it.Status,
		}
	}

	return resp
}

type UserUpdateRequest struct {
	Name     string `json:"name,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
	Role     string `json:"role,omitempty"`
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ToUserUpdateRequest(up UserUpdateRequest) domain.UpdateUser {
	return domain.UpdateUser{
		Name:         strPtr(up.Name),
		PasswordHash: strPtr(up.Password),
		Email:        strPtr(up.Email),
		Role:         strPtr(up.Role),
	}
}

//-------------------------------------------------------------------------------------------------

type StoreRequest struct {
	Name string `json:"name"`
}

type StoreResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func ToStoreResponse(store []domain.Store) []StoreResponse {
	resp := make([]StoreResponse, len(store))

	for i, s := range store {
		resp[i] = StoreResponse{
			ID:   s.ID,
			Name: s.Name,
		}
	}
	return resp
}

//-------------------------------------------------------------------------------------------------

type CategoryRequest struct {
	Name string `json:"name"`
}

type CategoryCreateResponse struct {
	CategoryID int `json:"categoryId"`
}

type CategoryResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func ToCategoryResponse(categories []domain.Category) []CategoryResponse {
	resp := make([]CategoryResponse, len(categories))

	for i, it := range categories {
		resp[i] = CategoryResponse{
			ID:   it.ID,
			Name: it.Name,
		}
	}

	return resp
}

//-------------------------------------------------------------------------------------------------

type ProductRequest struct {
	Title      string `json:"title"`
	Unit       string `json:"unit"`
	CategoryID int    `json:"categoryId"`
}

type ProductCreateResponse struct {
	ProductID int `json:"productId"`
}

type ProductResponse struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Unit     string `json:"unit"`
	Category string `json:"category"`
}

func ToProductResponse(product domain.ProductDetails) ProductResponse {
	return ProductResponse{
		ID:       product.ID,
		Title:    product.Title,
		Unit:     product.Unit,
		Category: product.Category,
	}
}

func ToProductsResponse(products []domain.ProductDetails) []ProductResponse {
	resp := make([]ProductResponse, len(products))

	for i, p := range products {
		resp[i] = ProductResponse{
			ID:       p.ID,
			Title:    p.Title,
			Unit:     p.Unit,
			Category: p.Category,
		}
	}
	return resp
}

//-------------------------------------------------------------------------------------------------

type ProductAliasRequest struct {
	Alias string `json:"alias"`
}

type ProductAliasCreateResponse struct {
	AliasID int `json:"aliasId"`
}
type ProductFindResponse struct {
	ProductID int `json:"productId"`
}

type ProductAliasResponse struct {
	ID      int    `json:"id"`
	Product string `json:"product"`
	Alias   string `json:"alias"`
}

func ToProductAliasResponse(alias domain.ProductAliasDetails) ProductAliasResponse {
	return ProductAliasResponse{
		ID:      alias.ID,
		Product: alias.Product,
		Alias:   alias.Alias,
	}
}

func ToProductAliasesResponse(alias []domain.ProductAliasDetails) []ProductAliasResponse {
	resp := make([]ProductAliasResponse, len(alias))

	for i, it := range alias {
		resp[i] = ProductAliasResponse{
			ID:      it.ID,
			Product: it.Product,
			Alias:   it.Alias,
		}
	}
	return resp
}

//-------------------------------------------------------------------------------------------------

type OrderRequest struct {
	StoreID int `json:"storeId"`
}

type OrderCreateResponse struct {
	ID int `json:"id"`
}

type OrderDetailsResponse struct {
	ID         int       `json:"id"`
	User       string    `json:"user"`
	Store      string    `json:"store"`
	ItemsCount int       `json:"itemsCount"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type OrderWithItemDetailsResponse struct {
	ID         int                   `json:"id"`
	User       string                `json:"user"`
	Store      string                `json:"store"`
	ItemsCount int                   `json:"itemsCount"`
	CreatedAt  time.Time             `json:"createdAt"`
	UpdatedAt  time.Time             `json:"updatedAt"`
	Items      []ItemDetailsResponse `json:"items"`
}

type ItemRequest struct {
	ProductID int `json:"productId"`
	Quantity  int `json:"quantity"`
}

type ItemUpdateRequest struct {
	Quantity int `json:"quantity"`
}

type ItemDetailsResponse struct {
	ID        int    `json:"id"`
	ProductID int    `json:"productId"`
	Title     string `json:"title"`
	Quantity  int    `json:"quantity"`
}

func ToResponseOrder(o domain.OrderWithItemDetails) OrderWithItemDetailsResponse {
	items := make([]ItemDetailsResponse, len(o.Items))
	for i, it := range o.Items {
		items[i] = ItemDetailsResponse{
			ProductID: it.ProductID,
			Title:     it.Title,
			Quantity:  it.Quantity,
		}
	}

	return OrderWithItemDetailsResponse{
		ID:         o.Order.ID,
		User:       o.Order.User,
		Store:      o.Order.Store,
		ItemsCount: o.Order.ItemsCount,
		CreatedAt:  o.Order.CreatedAt,
		UpdatedAt:  o.Order.UpdatedAt,
		Items:      items,
	}
}

func ToOrderListResponse(order []domain.OrderDetails) []OrderDetailsResponse {
	resp := make([]OrderDetailsResponse, len(order))

	for i, o := range order {
		resp[i] = OrderDetailsResponse{
			ID:         o.ID,
			User:       o.User,
			Store:      o.Store,
			ItemsCount: o.ItemsCount,
			CreatedAt:  o.CreatedAt,
			UpdatedAt:  o.UpdatedAt,
		}
	}
	return resp
}

//-------------------------------------------------------------------------------------------------
