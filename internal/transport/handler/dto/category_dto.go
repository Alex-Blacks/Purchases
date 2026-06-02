package dto

import "github.com/Alex-Blacks/Purchases/internal/domain"

type CategoryRequest struct {
	Name string `json:"name"`
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
