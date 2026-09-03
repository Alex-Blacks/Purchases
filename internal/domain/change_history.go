package domain

import (
	"context"
	"encoding/json"
	"time"
)

type HistoryEntity string
type HistoryAction string

const (
	HistoryEntityStore        HistoryEntity = "store"
	HistoryEntityUnit         HistoryEntity = "unit"
	HistoryEntityProduct      HistoryEntity = "product"
	HistoryEntityProductAlias HistoryEntity = "productAlias"
	HistoryEntityOrder        HistoryEntity = "order"
	HistoryEntityOrderItem    HistoryEntity = "orderItem"

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
	OldData    json.RawMessage
	NewData    json.RawMessage
	CreatedAt  time.Time
}

type HistoryListFilter struct {
	GroupIDs   []int
	EntityType *HistoryEntity
	EntityID   *int
	Action     *HistoryAction
	From       *time.Time
	To         *time.Time
	Limit      int
	Offset     int
}

type ChangeHistoryRepository interface {
	Insert(ctx context.Context, q Querier, groupID int, userID int, entityType HistoryEntity, entityID int, action HistoryAction, oldData json.RawMessage, newData json.RawMessage) error
	List(ctx context.Context, q Querier, filter HistoryListFilter) ([]ChangeHistoryDetails, error)
	Count(ctx context.Context, q Querier, filter HistoryListFilter) (int, error)
}
