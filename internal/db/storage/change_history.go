package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

type ChangeHistoryRepo struct{}

func NewHistoryRepo() *ChangeHistoryRepo {
	return &ChangeHistoryRepo{}
}

func (c *ChangeHistoryRepo) Insert(ctx context.Context, q domain.Querier, groupID int, userID int, entityType domain.HistoryEntity, entityID int, action domain.HistoryAction, oldData json.RawMessage, newData json.RawMessage) error {
	if _, err := q.Exec(ctx, `
		INSERT INTO change_history(group_id, user_id, entity_type, entity_id, action, old_data, new_data) VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, groupID, userID, entityType, entityID, action, oldData, newData); err != nil {
		return fmt.Errorf("query change history: %w", err)
	}
	return nil
}

func (c *ChangeHistoryRepo) List(ctx context.Context, q domain.Querier, filter domain.HistoryListFilter) ([]domain.ChangeHistoryDetails, error) {
	query := `
		SELECT h.id, h.group_id, g.name, h.user_id, u.name, h.entity_type, h.entity_id, h.action, h.old_data, h.new_data, h.created_at
		FROM change_history h
		JOIN groups g ON h.group_id = g.id
		JOIN users u ON h.user_id = u.id
	`
	args := []any{}
	setParts := []string{}
	argPos := 1

	if len(filter.GroupIDs) > 0 {
		setParts = append(setParts, fmt.Sprintf("WHERE h.group_id = ANY($%d::int[])", argPos))
		args = append(args, filter.GroupIDs)
		argPos++
	}

	if len(setParts) > 0 {
		query += " WHERE " + strings.Join(setParts, " AND ")
	}

	query += fmt.Sprintf("ORDER BY h.created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query change histories: %w", err)
	}

	defer rows.Close()
	var histories []domain.ChangeHistoryDetails
	for rows.Next() {
		var history domain.ChangeHistoryDetails
		if err := rows.Scan(&history.ID, &history.GroupID, &history.GroupName, &history.UserID, &history.UserName, &history.EntityType, &history.EntityID, &history.Action, &history.OldData, &history.NewData, &history.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan change history: %w", err)
		}

		histories = append(histories, history)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration failed: %w", err)
	}
	return histories, nil
}

func (c *ChangeHistoryRepo) Count(ctx context.Context, q domain.Querier, filter domain.HistoryListFilter) (int, error) {
	query := "SELECT COUNT(*) FROM change_history h"
	args := []any{}
	setParts := []string{}
	argPos := 1

	if len(filter.GroupIDs) > 0 {
		setParts = append(setParts, fmt.Sprintf("WHERE h.group_id = ANY($%d::int[])", argPos))
		args = append(args, filter.GroupIDs)
		argPos++
	}

	if len(setParts) > 0 {
		query += " WHERE " + strings.Join(setParts, " AND ")
	}

	var count int
	err := q.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}
