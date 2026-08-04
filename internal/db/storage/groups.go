package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

type GroupRepo struct{}

func NewGroupRepo() *GroupRepo {
	return &GroupRepo{}
}

func (g *GroupRepo) CreateGroup(ctx context.Context, q domain.Querier, name string, user_id int) (domain.GroupDetails, error) {
	var group domain.GroupDetails
	if err := q.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO groups(name, admin_user_id) 
			VALUES ($1,$2)
			RETURNING id, name, admin_user_id
		)
		SELECT i.id, i.name, i.admin_user_id, u.name
		FROM inserted i
		JOIN users u ON i.admin_user_id = u.id
	`, name, user_id).Scan(&group.Id, &group.Name, &group.Admin_user_id, &group.Admin_user); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.GroupDetails{}, domain.ErrAlreadyExists
		}
		return domain.GroupDetails{}, fmt.Errorf("query create group: %w", err)
	}
	return group, nil
}
