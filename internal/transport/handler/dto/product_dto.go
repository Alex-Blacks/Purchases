package dto

import "github.com/Alex-Blacks/Purchases/internal/domain"

type ProductRequest struct {
	Title      string `json:"title"`
	Unit       string `json:"unit"`
	CategoryID int    `json:"categoryId"`
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
	Product string `json:"product"`
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
