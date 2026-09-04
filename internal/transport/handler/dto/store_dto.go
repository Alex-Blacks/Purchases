package dto

import "github.com/Alex-Blacks/Purchases/internal/domain"

// StoreRequest используется для создания магазина.
type StoreRequest struct {
	Name    string `json:"name" validate:"required,min=1,max=50"`
	GroupID *int   `json:"groupId,omitempty" validate:"gt=0"` // опционально, для админов
}

// StoreUpdateRequest используется для обновления названия магазина.
type StoreUpdateRequest struct {
	Name *string `json:"name,omitempty" validate:"min=1,max=50"`
}

// StoreResponse возвращает информацию о магазине.
type StoreResponse struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	GroupID int    `json:"groupId"`
	Group   string `json:"group"` // название группы
}

// StoreFilterQuery – структура для биндинга query-параметров
type StoreFilterQuery struct {
	GroupIDs []int   `form:"group_ids" validate:"dive,int,gt=0"`
	Name     *string `form:"name,omitempty" validate:"min=1,max=50"`
	Limit    int     `form:"limit" default:"10" validate:"required,min=1,max=100"`
	Offset   int     `form:"offset" default:"0" validate:"min=0"`
}

// ToDomainFilter преобразует dto.UserFilterRequest в domain.UserListFilter.
func (q StoreFilterQuery) ToDomainFilter() domain.StoreListFilter {
	return domain.StoreListFilter{
		GroupIDs: q.GroupIDs,
		Name:     q.Name,
		Limit:    q.Limit,
		Offset:   q.Offset,
	}
}

// ToStoreResponse преобразует domain.StoreDetails в StoreResponse.
func ToStoreResponse(store domain.StoreDetails) StoreResponse {
	return StoreResponse{
		ID:      store.ID,
		Name:    store.Name,
		GroupID: store.GroupID,
		Group:   store.Group,
	}
}

// ToStoreUpdateRequest преобразует dto.StoreUpdateRequest в domain.StoreUpdate.
func ToStoreUpdateRequest(up StoreUpdateRequest) domain.StoreUpdate {
	return domain.StoreUpdate{
		Name: up.Name,
	}
}

// ToStoreListResponse преобразует слайс domain.StoreDetails в слайс StoreResponse.
func ToStoreListResponse(stores []domain.StoreDetails) []StoreResponse {
	resp := make([]StoreResponse, len(stores))
	for i, s := range stores {
		resp[i] = StoreResponse{
			ID:      s.ID,
			Name:    s.Name,
			GroupID: s.GroupID,
			Group:   s.Group,
		}
	}
	return resp
}
