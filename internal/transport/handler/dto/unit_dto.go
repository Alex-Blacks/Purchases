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

func ToUnitResponse(units []domain.Unit) []UnitResponse {
	resp := make([]UnitResponse, len(units))

	for i, u := range units {
		resp[i] = UnitResponse{
			Id:        u.Id,
			Name:      u.Name,
			ShortName: u.ShortName,
		}
	}
	return resp
}
