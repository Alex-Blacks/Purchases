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

type InviteRepo struct{}

func NewInviteRepo() *InviteRepo {
	return &InviteRepo{}
}

func (i *InviteRepo) Create(ctx context.Context, q domain.Querier, groupID int, inviterUserID int, inviteeEmail string, token string) (domain.InviteDetails, error) {
	var invite domain.InviteDetails
	if err := q.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO invites(group_id, inviter_user_id, invitee_email, token) 
			VALUES ($1,$2,$3,$4)
			RETURNING id, group_id, inviter_user_id, invitee_email, status, token, created_at, expires_at
		)
		SELECT i.id, i.group_id, g.name, i.inviter_user_id, u.name, i.invitee_email, i.status, i.token, i.created_at, i.expires_at
		FROM inserted i
		JOIN groups g ON i.group_id = g.id
		JOIN users u ON i.inviter_user_id = u.id
	`, groupID, inviterUserID, inviteeEmail, token).Scan(
		&invite.ID,
		&invite.GroupID,
		&invite.Group,
		&invite.InviterUserID,
		&invite.InviterUser,
		&invite.InviteeEmail,
		&invite.Status,
		&invite.Token,
		&invite.CreatedAt,
		&invite.ExpiresAt,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolation:
				return domain.InviteDetails{}, domain.ErrAlreadyExists
			case pgForeignKeyViolation:
				return domain.InviteDetails{}, domain.ErrConflict
			}
		}
		return domain.InviteDetails{}, fmt.Errorf("create invite: %w", err)
	}

	return invite, nil
}

func (i *InviteRepo) GetByID(ctx context.Context, q domain.Querier, inviteID int) (domain.InviteDetails, error) {
	var invite domain.InviteDetails
	if err := q.QueryRow(ctx, `
		SELECT i.id, i.group_id, g.name, i.inviter_user_id, u.name, i.invitee_email, i.status, i.token, i.created_at, i.expires_at
		FROM invites i
		JOIN groups g ON i.group_id = g.id
		JOIN users u ON i.inviter_user_id = u.id
		WHERE i.id = $1 AND
	`, inviteID).Scan(
		&invite.ID,
		&invite.GroupID,
		&invite.Group,
		&invite.InviterUserID,
		&invite.InviterUser,
		&invite.InviteeEmail,
		&invite.Status,
		&invite.Token,
		&invite.CreatedAt,
		&invite.ExpiresAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.InviteDetails{}, domain.ErrNotFound
		}
		return domain.InviteDetails{}, fmt.Errorf("get invite: %w", err)
	}

	return invite, nil
}

func (i *InviteRepo) GetByToken(ctx context.Context, q domain.Querier, token string) (domain.InviteDetails, error) {
	var invite domain.InviteDetails
	if err := q.QueryRow(ctx, `
		SELECT i.id, i.group_id, g.name, i.inviter_user_id, u.name, i.invitee_email, i.status, i.token, i.created_at, i.expires_at
		FROM invites i
		JOIN groups g ON i.group_id = g.id
		JOIN users u ON i.inviter_user_id = u.id
		WHERE i.token = $1
		ORDER BY created_at DESC LIMIT 1
	`, token).Scan(
		&invite.ID,
		&invite.GroupID,
		&invite.Group,
		&invite.InviterUserID,
		&invite.InviterUser,
		&invite.InviteeEmail,
		&invite.Status,
		&invite.Token,
		&invite.CreatedAt,
		&invite.ExpiresAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.InviteDetails{}, domain.ErrNotFound
		}
		return domain.InviteDetails{}, fmt.Errorf("get invite: %w", err)
	}

	return invite, nil
}

func (i *InviteRepo) GetByEmail(ctx context.Context, q domain.Querier, groupID int, email string) (domain.InviteDetails, error) {
	var invite domain.InviteDetails
	if err := q.QueryRow(ctx, `
		SELECT i.id, i.group_id, g.name, i.inviter_user_id, u.name, i.invitee_email, i.status, i.token, i.created_at, i.expires_at
		FROM invites i
		JOIN groups g ON i.group_id = g.id
		JOIN users u ON i.inviter_user_id = u.id
		WHERE i.invitee_email = $1 AND i.group_id = $2
		ORDER BY created_at DESC LIMIT 1
	`, email, groupID).Scan(
		&invite.ID,
		&invite.GroupID,
		&invite.Group,
		&invite.InviterUserID,
		&invite.InviterUser,
		&invite.InviteeEmail,
		&invite.Status,
		&invite.Token,
		&invite.CreatedAt,
		&invite.ExpiresAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.InviteDetails{}, domain.ErrNotFound
		}
		return domain.InviteDetails{}, fmt.Errorf("get invite: %w", err)
	}

	return invite, nil
}

func (i *InviteRepo) UpdateByID(ctx context.Context, q domain.Querier, inviteID int, groupID int, updateInvite domain.InviteUpdate) (domain.InviteDetails, error) {
	var invite domain.InviteDetails
	args := []any{inviteID, groupID}
	setPath := []string{}
	argPos := 3
	if updateInvite.Status != nil && *updateInvite.Status != "" {
		setPath = append(setPath, fmt.Sprintf("status = $%d", argPos))
		args = append(args, *updateInvite.Status)
		argPos++
	}
	if updateInvite.Token != nil && strings.TrimSpace(*updateInvite.Token) != "" {
		setPath = append(setPath, fmt.Sprintf("token = $%d", argPos))
		args = append(args, *updateInvite.Token)
		argPos++
	}

	set := strings.Join(setPath, ", ")
	if strings.TrimSpace(set) == "" {
		return domain.InviteDetails{}, domain.ErrNoFieldsToUpdate
	}
	if err := q.QueryRow(ctx, `
		UPDATE invites i
		SET `+set+`
		FROM groups g
		JOIN users u ON i.inviter_user_id = u.id
		WHERE i.id = $1 AND i.group_id = $2
		RETURNING i.id, i.group_id, g.name, i.inviter_user_id, u.name, i.invitee_email, i.status, i.token, i.created_at, i.expires_at
	`, args...).Scan(
		&invite.ID,
		&invite.GroupID,
		&invite.Group,
		&invite.InviterUserID,
		&invite.InviterUser,
		&invite.InviteeEmail,
		&invite.Status,
		&invite.Token,
		&invite.CreatedAt,
		&invite.ExpiresAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.InviteDetails{}, domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolation:
				return domain.InviteDetails{}, domain.ErrAlreadyExists
			case pgForeignKeyViolation:
				return domain.InviteDetails{}, domain.ErrConflict
			}
		}
		return domain.InviteDetails{}, fmt.Errorf("update invite: %w", err)
	}

	return invite, nil
}

func (i *InviteRepo) DeleteByID(ctx context.Context, q domain.Querier, inviteID int, groupID int) error {
	var id int
	if err := q.QueryRow(ctx, `DELETE FROM invites WHERE id = $1 AND group_id = $2 RETURNING id`, inviteID, groupID).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
			return domain.ErrConflict
		}
		return fmt.Errorf("delete invite: %w", err)
	}

	return nil
}

func (i *InviteRepo) List(ctx context.Context, q domain.Querier, groupID int) ([]domain.InviteDetails, error) {
	row, err := q.Query(ctx, `
		SELECT i.id, i.group_id, g.name, i.inviter_user_id, u.name, i.invitee_email, i.status, i.token, i.created_at, i.expires_at
		FROM invites i
		JOIN groups g ON i.group_id = g.id
		JOIN users u ON i.inviter_user_id = u.id
		WHERE i.group_id = $1
		ORDER BY created_at DESC
	`, groupID)
	if err != nil {
		return []domain.InviteDetails{}, fmt.Errorf("query invites: %w", err)
	}

	var invites []domain.InviteDetails
	for row.Next() {
		var invite domain.InviteDetails
		if err := row.Scan(
			&invite.ID, &invite.GroupID, &invite.Group, &invite.InviterUserID, &invite.InviterUser, &invite.InviteeEmail, &invite.Status, &invite.Token, &invite.CreatedAt, &invite.ExpiresAt,
		); err != nil {
			return []domain.InviteDetails{}, fmt.Errorf("scan invite: %w", err)
		}

		invites = append(invites, invite)
	}

	if err := row.Err(); err != nil {
		return []domain.InviteDetails{}, fmt.Errorf("iteration failed: %w", err)
	}

	return invites, nil
}

func (i *InviteRepo) ListAll(ctx context.Context, q domain.Querier) ([]domain.InviteDetails, error) {
	row, err := q.Query(ctx, `
		SELECT i.id, i.group_id, g.name, i.inviter_user_id, u.name, i.invitee_email, i.status, i.token, i.created_at, i.expires_at
		FROM invites i
		JOIN groups g ON i.group_id = g.id
		JOIN users u ON i.inviter_user_id = u.id
		ORDER BY created_at DESC
	`)
	if err != nil {
		return []domain.InviteDetails{}, fmt.Errorf("query invites: %w", err)
	}

	var invites []domain.InviteDetails
	for row.Next() {
		var invite domain.InviteDetails
		if err := row.Scan(
			&invite.ID, &invite.GroupID, &invite.Group, &invite.InviterUserID, &invite.InviterUser, &invite.InviteeEmail, &invite.Status, &invite.Token, &invite.CreatedAt, &invite.ExpiresAt,
		); err != nil {
			return []domain.InviteDetails{}, fmt.Errorf("scan invite: %w", err)
		}

		invites = append(invites, invite)
	}

	if err := row.Err(); err != nil {
		return []domain.InviteDetails{}, fmt.Errorf("iteration failed: %w", err)
	}

	return invites, nil
}
