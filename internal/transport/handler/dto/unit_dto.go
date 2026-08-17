package dto

import "github.com/Alex-Blacks/Purchases/internal/domain"

type UnitRequest struct {
	Name      string `json:"name" validate:"required,min=1,max=50"`
	ShortName string `json:"shortName" validate:"required,min=1,max=50"`
}

type UnitUpdateRequest struct {
	Name      *string `json:"name,omitempty" validate:"min=1,max=50"`
	ShortName *string `json:"shortName,omitempty" validate:"min=1,max=50"`
}

type UnitResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	GroupID   int    `json:"groupId"`
	Group     string `json:"group"`
}

func ToUnitResponse(unit domain.UnitDetails) UnitResponse {
	return UnitResponse{
		ID:        unit.ID,
		Name:      unit.Name,
		ShortName: unit.ShortName,
		GroupID:   unit.GroupID,
		Group:     unit.Group,
	}
}

func ToUnitUpdateRequest(up UnitUpdateRequest) domain.UnitUpdate {
	return domain.UnitUpdate{
		Name:      up.Name,
		ShortName: up.ShortName,
	}
}

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
