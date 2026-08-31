package domain

import (
	"context"
	"encoding/json"
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
	OldData    json.RawMessage
	NewData    json.RawMessage
	CreatedAt  time.Time
}

type HistoryListFilter struct {
	GroupIDs   []int          `json:"group_ids"`
	EntityType *HistoryEntity `json:"entity_type,omitempty"`
	EntityID   *int           `json:"entity_id,omitempty"`
	Action     *HistoryAction `json:"action,omitempty"`
	From       *time.Time     `json:"from,omitempty"`
	To         *time.Time     `json:"to,omitempty"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
}

type ChangeHistoryRepository interface {
	Insert(ctx context.Context, q Querier, groupID int, userID int, entityType HistoryEntity, entityID int, action HistoryAction, oldData json.RawMessage, newData json.RawMessage) error
	List(ctx context.Context, q Querier, filter HistoryListFilter) ([]ChangeHistoryDetails, error)
	Count(ctx context.Context, q Querier, filter HistoryListFilter) (int, error)
}
