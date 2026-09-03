package dto

import (
	"github.com/Alex-Blacks/Purchases/internal/domain"
)

// UnitRequest используется для создания единицы измерения.
type UnitRequest struct {
	Name      string `json:"name" validate:"required,min=1,max=50"`
	ShortName string `json:"shortName" validate:"required,min=1,max=50"`
	GroupID   *int   `json:"groupId,omitempty" validate:"gt=0"` // опционально, для админов
}

// UnitUpdateRequest используется для обновления единицы измерения.
type UnitUpdateRequest struct {
	Name      *string `json:"name,omitempty" validate:"min=1,max=50"`
	ShortName *string `json:"shortName,omitempty" validate:"min=1,max=50"`
}

// UnitResponse возвращает информацию о единице измерения.
type UnitResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	GroupID   int    `json:"groupId"`
	Group     string `json:"group"` // название группы
}

// UnitFilterQuery – структура для биндинга query-параметров
type UnitFilterQuery struct {
	GroupIDs  []int   `form:"group_ids" validate:"dive,int,gt=0"`
	Name      *string `form:"name,omitempty" validate:"min=1,max=50"`
	ShortName *string `form:"short_name,omitempty" validate:"min=1,max=50"`
	Limit     int     `form:"limit" default:"10" validate:"required,min=1,max=100"`
	Offset    int     `form:"offset" default:"0" validate:"min=0"`
}

// ToUserFilterRequest преобразует dto.UserFilterRequest в domain.UserListFilter.
func (q UnitFilterQuery) ToDomainFilter() domain.UnitListFilter {
	return domain.UnitListFilter{
		GroupIDs:  q.GroupIDs,
		Name:      q.Name,
		ShortName: q.ShortName,
		Limit:     q.Limit,
		Offset:    q.Offset,
	}
}

// ToUnitResponse преобразует domain.UnitDetails в UnitResponse.
func ToUnitResponse(unit domain.UnitDetails) UnitResponse {
	return UnitResponse{
		ID:        unit.ID,
		Name:      unit.Name,
		ShortName: unit.ShortName,
		GroupID:   unit.GroupID,
		Group:     unit.Group,
	}
}

// ToUnitUpdateRequest преобразует dto.UnitUpdateRequest в domain.UnitUpdate.
func ToUnitUpdateRequest(up UnitUpdateRequest) domain.UnitUpdate {
	return domain.UnitUpdate{
		Name:      up.Name,
		ShortName: up.ShortName,
	}
}

// ToListUnitResponse преобразует слайс domain.UnitDetails в слайс UnitResponse.
func ToListUnitResponse(units []domain.UnitDetails) []UnitResponse {
	resp := make([]UnitResponse, len(units))

	for i, u := range units {
		resp[i] = UnitResponse{
			ID:        u.ID,
			Name:      u.Name,
			ShortName: u.ShortName,
			GroupID:   u.GroupID,
			Group:     u.Group,
		}
	}
	return resp
}
