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

func (g *GroupRepo) Create(ctx context.Context, q domain.Querier, name string, adminUserID *int) (domain.GroupDetails, error) {
	var group domain.GroupDetails
	if err := q.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO groups(name, admin_user_id) 
			VALUES ($1,$2)
			RETURNING id, name, admin_user_id
		)
		SELECT i.id, i.name, i.admin_user_id, COALESCE(u.name, '')
		FROM inserted i
		LEFT JOIN users u ON i.admin_user_id = u.id
	`, name, adminUserID).Scan(&group.ID, &group.Name, &group.AdminUserID, &group.AdminUser); err != nil {
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

func (g *GroupRepo) GetByID(ctx context.Context, q domain.Querier, groupID int) (domain.GroupDetails, error) {
	var group domain.GroupDetails
	if err := q.QueryRow(ctx, `
		SELECT g.id, g.name, g.admin_user_id, COALESCE(u.name, '')
		FROM groups g
		LEFT JOIN users u ON g.admin_user_id = u.id
		WHERE g.id = $1
	`, groupID).Scan(&group.ID, &group.Name, &group.AdminUserID, &group.AdminUser); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.GroupDetails{}, domain.ErrNotFound
		}
		return domain.GroupDetails{}, fmt.Errorf("get group: %w", err)
	}
	return group, nil
}

func (g *GroupRepo) CheckGroupAdmin(ctx context.Context, q domain.Querier, groupID int, adminUserID int) (bool, error) {
	var id int
	if err := q.QueryRow(ctx, `SELECT 1 FROM groups WHERE id = $1 AND admin_user_id = $2`, groupID, adminUserID).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check group admin: %w", err)
	}
	return true, nil
}

func (g *GroupRepo) UpdateByID(ctx context.Context, q domain.Querier, groupID int, updateGroup domain.GroupUpdate) (domain.GroupDetails, error) {
	args := []any{groupID}
	setParts := []string{}
	argPos := 2
	if updateGroup.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argPos))
		args = append(args, *updateGroup.Name)
		argPos++
	}
	if updateGroup.AdminUserID != nil {
		setParts = append(setParts, fmt.Sprintf("admin_user_id = $%d", argPos))
		args = append(args, *updateGroup.AdminUserID)
		argPos++
	}

	set := strings.Join(setParts, ", ")
	if strings.TrimSpace(set) == "" {
		return domain.GroupDetails{}, domain.ErrNoFieldsToUpdate
	}

	var group domain.GroupDetails
	if err := q.QueryRow(ctx, `
		UPDATE groups g
		SET `+set+`
		WHERE g.id = $1
		RETURNING g.id, g.name, g.admin_user_id, (SELECT u.name FROM users u WHERE u.id = g.admin_user_id)
		
	`, args...).Scan(&group.ID, &group.Name, &group.AdminUserID, &group.AdminUser); err != nil {
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

func (g *GroupRepo) UpdateGroupAdmin(ctx context.Context, q domain.Querier, groupID, adminUserID int) error {
	tag, err := q.Exec(ctx, `UPDATE groups SET admin_user_id = $1 WHERE id = $2`, adminUserID, groupID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
			return domain.ErrConflict
		}
		return fmt.Errorf("update group admin: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (g *GroupRepo) DeleteByID(ctx context.Context, q domain.Querier, groupID int) error {
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

func (g *GroupRepo) List(ctx context.Context, q domain.Querier, filter domain.GroupListFilter) ([]domain.GroupDetails, error) {
	query := `
		SELECT g.id, g.name, g.admin_user_id, COALESCE(u.name, '')
		FROM groups g
		LEFT JOIN users u ON g.admin_user_id = u.id`

	whereClause, whereArgs, whereArgPos := buildGroupWhere(filter)
	query += whereClause

	query += fmt.Sprintf("ORDER BY g.id DESC LIMIT $%d OFFSET $%d", whereArgPos, whereArgPos+1)
	args := append(whereArgs, filter.Limit, filter.Offset)

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query groups: %w", err)
	}
	defer rows.Close()

	var groups []domain.GroupDetails
	for rows.Next() {
		var group domain.GroupDetails
		if err := rows.Scan(&group.ID, &group.Name, &group.AdminUserID, &group.AdminUser); err != nil {
			return nil, fmt.Errorf("scan group: %w", err)
		}

		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration failed: %w", err)
	}

	return groups, nil
}

func (g *GroupRepo) Count(ctx context.Context, q domain.Querier, filter domain.GroupListFilter) (int, error) {
	query := "SELECT COUNT(*) FROM groups g"

	whereClause, args, _ := buildGroupWhere(filter)
	query += whereClause

	var count int
	if err := q.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("query count group: %w", err)
	}

	return count, nil
}
