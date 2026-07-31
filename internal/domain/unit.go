package domain

import "context"

type Unit struct {
	Id        int
	Name      string
	ShortName string
}

type UnitRepository interface {
	ListUnits(ctx context.Context, q Querier) ([]Unit, error)
}
