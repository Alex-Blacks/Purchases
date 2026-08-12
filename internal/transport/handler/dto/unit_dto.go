package dto

import "github.com/Alex-Blacks/Purchases/internal/domain"

type UnitRequest struct {
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
}

type UnitResponse struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
}

func ToUnitResponse(units []domain.UnitDetails) []UnitResponse {
	resp := make([]UnitResponse, len(units))

	for i, u := range units {
		resp[i] = UnitResponse{
			Id:        u.ID,
			Name:      u.Name,
			ShortName: u.ShortName,
		}
	}
	return resp
}
