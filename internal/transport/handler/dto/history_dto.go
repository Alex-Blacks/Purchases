package dto

import (
	"encoding/json"
	"time"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

type ChangeHistoryResponse struct {
	ID         int                  `json:"id"`
	GroupID    int                  `json:"groupId"`
	GroupName  string               `json:"group"`
	UserID     int                  `json:"userId"`
	UserName   string               `json:"userName"`
	EntityType domain.HistoryEntity `json:"entity_type"`
	EntityID   int                  `json:"entityId"`
	Action     domain.HistoryAction `json:"action"`
	OldData    json.RawMessage      `json:"oldData"`
	NewData    json.RawMessage      `json:"newData"`
	CreatedAt  time.Time            `json:"createdAt"`
}

func ToChangeHistoryListResponse(histories []domain.ChangeHistoryDetails) []ChangeHistoryResponse {
	resp := make([]ChangeHistoryResponse, len(histories))

	for i, history := range histories {
		resp[i] = ChangeHistoryResponse{
			ID:         history.ID,
			GroupID:    history.GroupID,
			GroupName:  history.GroupName,
			UserID:     history.UserID,
			UserName:   history.UserName,
			EntityType: history.EntityType,
			EntityID:   history.EntityID,
			Action:     history.Action,
			OldData:    history.OldData,
			NewData:    history.NewData,
			CreatedAt:  history.CreatedAt,
		}
	}

	return resp
}

// HistoryFilterQuery – структура для биндинга query-параметров
type HistoryFilterQuery struct {
	GroupIDs   []int                 `form:"group_ids" validate:"dive,int,gt=0"`
	EntityType *domain.HistoryEntity `form:"entity_type,omitempty" validate:"oneof=store unit product productAlias order orderItem"`
	EntityID   *int                  `form:"entity_id,omitempty" validate:"gt=0"`
	Action     *domain.HistoryAction `form:"action,omitempty" validate:"oneof=create update delete"`
	From       *time.Time            `form:"from" validate:"daterange=To"`
	To         *time.Time            `form:"to"`
	Limit      int                   `form:"limit" default:"10" validate:"required,min=1,max=100"`
	Offset     int                   `form:"offset" default:"0" validate:"min=0"`
}

// ToDomainFilter преобразует dto.HistoryFilterQuery в domain.HistoryListFilter.
func (q HistoryFilterQuery) ToDomainFilter() domain.HistoryListFilter {
	return domain.HistoryListFilter{
		GroupIDs:   q.GroupIDs,
		EntityType: q.EntityType,
		EntityID:   q.EntityID,
		Action:     q.Action,
		From:       q.From,
		To:         q.To,
		Limit:      q.Limit,
		Offset:     q.Offset,
	}
}
