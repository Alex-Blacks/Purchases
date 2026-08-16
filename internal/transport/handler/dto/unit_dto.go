package dto

import "github.com/Alex-Blacks/Purchases/internal/domain"

type UnitRequest struct {
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
}

type UnitResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	GroupID   int    `json:"groupID"`
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
