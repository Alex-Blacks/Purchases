package dto

import "github.com/Alex-Blacks/Purchases/internal/domain"

// ProductRequest используется для создания продукта.
type ProductRequest struct {
	Title   string `json:"title" validate:"required,min=1,max=50"`
	GroupID *int   `json:"groupId,omitempty" validate:"gt=0"` // опционально, для админов
}

// ProductUpdateRequest используется для обновления продукта.
type ProductUpdateRequest struct {
	Title *string `json:"title" validate:"min=1,max=50"`
}

// ProductResponse возвращает информацию продукте.
type ProductResponse struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	GroupID int    `json:"groupId"`
	Group   string `json:"group"` // название группы
}

// ToProductResponse преобразует domain.ProductDetails в ProductResponse.
func ToProductResponse(product domain.ProductDetails) ProductResponse {
	return ProductResponse{
		ID:      product.ID,
		Title:   product.Title,
		GroupID: product.GroupID,
		Group:   product.Group,
	}
}

// ToProductUpdateRequest преобразует dto.ProductUpdateRequest в domain.ProductUpdate.
func ToProductUpdateRequest(product ProductUpdateRequest) domain.ProductUpdate {
	return domain.ProductUpdate{
		Title: product.Title,
	}
}

// ToListProductResponse преобразует слайс dto.ProductDetails в domain.ProductResponse.
func ToProductListResponse(products []domain.ProductDetails) []ProductResponse {
	resp := make([]ProductResponse, len(products))

	for i, p := range products {
		resp[i] = ProductResponse{
			ID:      p.ID,
			Title:   p.Title,
			GroupID: p.GroupID,
			Group:   p.Group,
		}
	}
	return resp
}

// ---------------------------------------------------------------------------------
// --------------------------------------ALIAS--------------------------------------
// ---------------------------------------------------------------------------------

// ProductAliasRequest используется для создания алиаса продукта.
type ProductAliasRequest struct {
	Alias   string `json:"alias" validate:"required,min=1,max=50"`
	GroupID *int   `json:"groupId,omitempty" validate:"gt=0"` // опционально, для админов
}

// ProductAliasUpdateRequest используется для обновления алиаса продукта.
type ProductAliasUpdateRequest struct {
	Alias *string `json:"alias" validate:"min=1,max=50"`
}

// ProductAliasResponse возвращает информацию об алиасе продукта.
type ProductAliasResponse struct {
	ID        int    `json:"id"`
	ProductID int    `json:"productId"`
	Product   string `json:"product"`
	Alias     string `json:"alias"`
	GroupID   int    `json:"groupId"`
	Group     string `json:"group"`
}

// ToProductAliasResponse преобразует domain.ProductAliasDetails в ProductAliasResponse.
func ToProductAliasResponse(alias domain.ProductAliasDetails) ProductAliasResponse {
	return ProductAliasResponse{
		ID:        alias.ID,
		ProductID: alias.ProductID,
		Product:   alias.Product,
		Alias:     alias.Alias,
		GroupID:   alias.GroupID,
		Group:     alias.Group,
	}
}

// ToProductAliasUpdateRequest преобразует domain.ProductAliasUpdateRequest в ProductAliasUpdate.
func ToProductAliasUpdateRequest(alias ProductAliasUpdateRequest) domain.ProductAliasUpdate {
	return domain.ProductAliasUpdate{
		Alias: alias.Alias,
	}
}

// ToProductAliasListResponse преобразует слайс dto.ProductAliasDetails в domain.ProductAliasResponse.
func ToProductAliasListResponse(alias []domain.ProductAliasDetails) []ProductAliasResponse {
	resp := make([]ProductAliasResponse, len(alias))

	for i, it := range alias {
		resp[i] = ProductAliasResponse{
			ID:        it.ID,
			ProductID: it.ProductID,
			Product:   it.Product,
			Alias:     it.Alias,
			GroupID:   it.GroupID,
			Group:     it.Group,
		}
	}
	return resp
}
