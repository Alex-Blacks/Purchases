package dto

import "github.com/Alex-Blacks/Purchases/internal/domain"

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
