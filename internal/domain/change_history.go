package domain

import (
	"context"
	"time"
)

type HistoryEntity string
type HistoryAction string

const (
	HistoryEntityStore   HistoryEntity = "store"
	HistoryEntityUnit    HistoryEntity = "unit"
	HistoryEntityProduct HistoryEntity = "product"
	HistoryEntityOrder   HistoryEntity = "order"

	HistoryActionCreate HistoryAction = "create"
	HistoryActionUpdate HistoryAction = "update"
	HistoryActionDelete HistoryAction = "delete"
)

type ChangeHistoryDetails struct {
	ID         int
	GroupID    int
	GroupName  string
	UserID     int
	UserName   string
	EntityType HistoryEntity
	EntityID   int
	Action     HistoryAction
	OldData    any
	NewData    any
	CreatedAt  time.Time
}

type ChangeHistoryRepository interface {
	Insert(ctx context.Context, q Querier, groupID, userID int, entityType HistoryEntity, entityID int, action HistoryAction, newData any) error
	List(ctx context.Context, q Querier, groupID int) ([]ChangeHistoryDetails, error)
}
