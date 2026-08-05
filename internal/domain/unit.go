package domain

import "context"

type UnitDetails struct {
	Id        int
	Name      string
	ShortName string
	GroupID   int
	Group     string
}

type UnitUpdate struct {
	Name      *string
	ShortName *string
	GroupID   *int
}

type UnitRepository interface {
	CreateUnit(ctx context.Context, q Querier, name string, groupID int, shortName string) (UnitDetails, error)
	GetUnitByID(ctx context.Context, q Querier, groupID int) (UnitDetails, error)
	UpdateUnitByID(ctx context.Context, q Querier, unitID int, unitUpdate UnitUpdate) (UnitDetails, error)
	DeleteUnitByID(ctx context.Context, q Querier, groupID int) error
	ListUnits(ctx context.Context, q Querier) ([]UnitDetails, error)
}
