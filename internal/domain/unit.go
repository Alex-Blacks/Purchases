package domain

import "context"

type UnitDetails struct {
	ID        int
	Name      string
	ShortName string
	GroupID   int
	Group     string
}

type UnitUpdate struct {
	Name      *string
	ShortName *string
}

func (u UnitDetails) GetGroupID() int { return u.GroupID }

type UnitRepository interface {
	CreateUnit(ctx context.Context, q Querier, name string, groupID int, shortName string) (UnitDetails, error)
	GetUnitByID(ctx context.Context, q Querier, unitID int) (UnitDetails, error)
	UpdateUnitByID(ctx context.Context, q Querier, unitID int, unitUpdate UnitUpdate) (UnitDetails, error)
	DeleteUnitByID(ctx context.Context, q Querier, unitID int) error
	ListUnits(ctx context.Context, q Querier, groupID []int) ([]UnitDetails, error)
	ListAdminUnits(ctx context.Context, q Querier) ([]UnitDetails, error)
}
