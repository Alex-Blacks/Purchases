package domain

import "context"

type UnitCreate struct {
	Name      string
	ShortName string
}

type UnitUpdate struct {
	Name      *string
	ShortName *string
}

type UnitDetails struct {
	ID        int
	Name      string
	ShortName string
	GroupID   int
	Group     string
}

func (u UnitDetails) GetGroupID() int { return u.GroupID }
func (u UnitDetails) GetID() int      { return u.ID }

type UnitRepository interface {
	Create(ctx context.Context, q Querier, params any, groupID int) (UnitDetails, error)
	GetByID(ctx context.Context, q Querier, id int) (UnitDetails, error)
	UpdateByID(ctx context.Context, q Querier, id int, updates any) (UnitDetails, error)
	DeleteByID(ctx context.Context, q Querier, id int) error
	List(ctx context.Context, q Querier, groupIDs []int) ([]UnitDetails, error)
	ListAll(ctx context.Context, q Querier) ([]UnitDetails, error)
}
