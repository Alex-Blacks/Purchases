package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type GroupRepo struct{}

func NewGroupRepo() *GroupRepo {
	return &GroupRepo{}
}

func (g *GroupRepo) CreateGroup(ctx context.Context, q domain.Querier, name string, userID int) (domain.GroupDetails, error) {
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
	`, name, userID).Scan(&group.Id, &group.Name, &group.AdminUserID, &group.AdminUser); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolation:
				return domain.GroupDetails{}, domain.ErrAlreadyExists
			case pgForeignKeyViolation:
				return domain.GroupDetails{}, domain.ErrConflict
			}
		}
		return domain.GroupDetails{}, fmt.Errorf("query create group: %w", err)
	}
	return group, nil
}

func (g *GroupRepo) GetGroupById(ctx context.Context, q domain.Querier, groupID int) (domain.GroupDetails, error) {
	var group domain.GroupDetails
	if err := q.QueryRow(ctx, `
		SELECT g.id, g.name, g.admin_user_id, u.name
		FROM groups g
		JOIN users u ON g.admin_user_id = u.id
		WHERE g.id = $1
	`, groupID).Scan(&group.Id, &group.Name, &group.AdminUserID, &group.AdminUser); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.GroupDetails{}, domain.ErrNotFound
		}
		return domain.GroupDetails{}, fmt.Errorf("get group: %w", err)
	}
	return group, nil
}

func (g *GroupRepo) UpdateGroupByID(ctx context.Context, q domain.Querier, groupID int, updateGroup domain.GroupUpdate) (domain.GroupDetails, error) {
	var group domain.GroupDetails
	args := []any{groupID}
	setParts := []string{}
	argPos := 2
	if updateGroup.Name != nil && strings.TrimSpace(*updateGroup.Name) != "" {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argPos))
		args = append(args, *updateGroup.Name)
		argPos++
	}
	if updateGroup.AdminUserID != nil && *updateGroup.AdminUserID >= 1 {
		setParts = append(setParts, fmt.Sprintf("admin_user_id = $%d", argPos))
		args = append(args, *updateGroup.AdminUserID)
		argPos++
	}

	set := strings.Join(setParts, ", ")
	if strings.TrimSpace(set) == "" {
		return domain.GroupDetails{}, domain.ErrNoFieldsToUpdate
	}
	if err := q.QueryRow(ctx, `
		UPDATE groups g
		SET `+set+`
		FROM users u
		WHERE g.id = $1 AND g.admin_user_id = u.id
		RETURNING g.id, g.name, g.admin_user_id, u.name
	`, args...).Scan(&group.Id, &group.Name, &group.AdminUserID, &group.AdminUser); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.GroupDetails{}, domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolation:
				return domain.GroupDetails{}, domain.ErrAlreadyExists
			case pgForeignKeyViolation:
				return domain.GroupDetails{}, domain.ErrConflict
			}
		}
		return domain.GroupDetails{}, fmt.Errorf("update group: %w", err)
	}
	return group, nil
}

func (g *GroupRepo) DeleteGroupByID(ctx context.Context, q domain.Querier, groupID int) error {
	var id int
	if err := q.QueryRow(ctx, `DELETE FROM groups WHERE groups.id = $1 RETURNING id`, groupID).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
			return domain.ErrConflict
		}
		return fmt.Errorf("delete group: %w", err)
	}
	return nil
}

func (g *GroupRepo) ListGroups(ctx context.Context, q domain.Querier) ([]domain.GroupDetails, error) {
	rows, err := q.Query(ctx, `
		SELECT g.id, g.name, g.admin_user_id, u.name
		FROM groups g
		JOIN users u ON g.admin_user_id = u.id
	`)
	if err != nil {
		return []domain.GroupDetails{}, fmt.Errorf("query groups: %w", err)
	}

	var groups []domain.GroupDetails
	for rows.Next() {
		var group domain.GroupDetails
		if err := rows.Scan(&group.Id, &group.Name, &group.AdminUserID, &group.AdminUser); err != nil {
			return []domain.GroupDetails{}, fmt.Errorf("scan group: %w", err)
		}

		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return []domain.GroupDetails{}, fmt.Errorf("iteration failed: %w", err)
	}

	return groups, nil
}
