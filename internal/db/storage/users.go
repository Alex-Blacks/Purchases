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

type UserRepo struct{}

func NewUserRepo() *UserRepo {
	return &UserRepo{}
}

func (u *UserRepo) CreateUser(ctx context.Context, q domain.Querier, name, passwordHash, email string, groupID int, role, status string) (domain.UserDetails, error) {
	var user domain.UserDetails
	if err := q.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO users(name, password_hash, email, group_id, role, status) 
			VALUES ($1,$2,$3,$4,$5,$6) 
			RETURNING id, name, password_hash, email, group_id, role, status
		)
		SELECT i.id, i.name, i.password_hash, i.email, i.group_id, g.name, i.role, i.status
		FROM inserted i
		JOIN groups g ON i.group_id = g.id
	`, name, passwordHash, email, groupID, role, status).Scan(&user.ID, &user.Name, &user.PasswordHash, &user.Email, &user.GroupID, &user.Group, &user.Role, &user.Status); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolation:
				return domain.UserDetails{}, domain.ErrAlreadyExists
			case pgForeignKeyViolation:
				return domain.UserDetails{}, domain.ErrConflict
			}
		}
		return domain.UserDetails{}, fmt.Errorf("query create user: %w", err)
	}
	return user, nil
}
func (u *UserRepo) GetUserByID(ctx context.Context, q domain.Querier, userID int) (domain.UserDetails, error) {
	var user domain.UserDetails
	if err := q.QueryRow(ctx, `
		SELECT u.id, u.name, u.password_hash, u.email, u.group_id, g.name, u.role, u.status 
		FROM users u
		JOIN groups g ON u.group_id = g.id
		WHERE u.id = $1
	`, userID).Scan(&user.ID, &user.Name, &user.PasswordHash, &user.Email, &user.GroupID, &user.Group, &user.Role, &user.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.UserDetails{}, domain.ErrNotFound
		}
		return domain.UserDetails{}, fmt.Errorf("query get user: %w", err)
	}
	return user, nil
}

func (u *UserRepo) UpdateUserByID(ctx context.Context, q domain.Querier, userID int, groupID int, updateUser domain.UserUpdate) (domain.UserDetails, error) {
	var user domain.UserDetails
	args := []any{userID, groupID}
	setParts := []string{}
	argPos := 3
	if updateUser.Name != nil && strings.TrimSpace(*updateUser.Name) != "" {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argPos))
		args = append(args, *updateUser.Name)
		argPos++
	}
	if updateUser.Password != nil && strings.TrimSpace(*updateUser.Password) != "" {
		setParts = append(setParts, fmt.Sprintf("password_hash = $%d", argPos))
		args = append(args, *updateUser.Password)
		argPos++
	}
	if updateUser.Email != nil && strings.TrimSpace(*updateUser.Email) != "" {
		setParts = append(setParts, fmt.Sprintf("email = $%d", argPos))
		args = append(args, *updateUser.Email)
		argPos++
	}
	if updateUser.GroupID != nil && *updateUser.GroupID >= 1 {
		setParts = append(setParts, fmt.Sprintf("group_id = $%d", argPos))
		args = append(args, *updateUser.GroupID)
		argPos++
	}
	if updateUser.Role != nil && strings.TrimSpace(*updateUser.Role) != "" {
		setParts = append(setParts, fmt.Sprintf("role = $%d", argPos))
		args = append(args, *updateUser.Role)
		argPos++
	}
	if updateUser.Status != nil && strings.TrimSpace(*updateUser.Status) != "" {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argPos))
		args = append(args, *updateUser.Status)
		argPos++
	}

	set := strings.Join(setParts, ", ")
	if strings.TrimSpace(set) == "" {
		return domain.UserDetails{}, domain.ErrNoFieldsToUpdate
	}
	if err := q.QueryRow(ctx, `
		UPDATE users u
		SET `+set+`
		FROM groups g
		WHERE u.id = $1 AND u.group_id = $2 AND u.group_id = g.id
		RETURNING u.id, u.name, u.password_hash, u.email, u.group_id, g.name, u.role, u.status
	`, args...).Scan(&user.ID, &user.Name, &user.PasswordHash, &user.Email, &user.GroupID, &user.Group, &user.Role, &user.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.UserDetails{}, domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolation:
				return domain.UserDetails{}, domain.ErrAlreadyExists
			case pgForeignKeyViolation:
				return domain.UserDetails{}, domain.ErrConflict
			}
		}
		return domain.UserDetails{}, fmt.Errorf("query get user: %w", err)
	}
	return user, nil
}

func (u *UserRepo) DeleteUserByID(ctx context.Context, q domain.Querier, userID int, groupID int) error {
	var id int
	if err := q.QueryRow(ctx, `DELETE FROM users WHERE users.id = $1 AND users.group_id = $2 RETURNING id`, userID, groupID).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
			return domain.ErrConflict
		}
		return fmt.Errorf("query delete user: %w", err)
	}
	return nil
}
func (u *UserRepo) ListUsers(ctx context.Context, q domain.Querier, groupID int) ([]domain.UserDetails, error) {
	rows, err := q.Query(ctx, `
		SELECT u.id, u.name, u.password_hash, u.email, u.group_id, g.name, u.role, u.status 
		FROM users u
		JOIN groups g ON u.group_id = g.id
		WHERE u.group_id = $1
	`, groupID)
	if err != nil {
		return []domain.UserDetails{}, fmt.Errorf("query list users: %w", err)
	}
	defer rows.Close()

	var users []domain.UserDetails
	for rows.Next() {
		var user domain.UserDetails

		if err := rows.Scan(&user.ID, &user.Name, &user.PasswordHash, &user.Email, &user.GroupID, &user.Group, &user.Role, &user.Status); err != nil {
			return []domain.UserDetails{}, fmt.Errorf("scan list users: %w", err)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return []domain.UserDetails{}, fmt.Errorf("iteration failed: %w", err)
	}
	return users, nil
}

func (u *UserRepo) GetUserByEmail(ctx context.Context, q domain.Querier, email string) (domain.UserDetails, error) {
	var user domain.UserDetails
	if err := q.QueryRow(ctx, `
		SELECT u.id, u.name, u.password_hash, u.email, u.group_id, g.name, u.role, u.status 
		FROM users u
		JOIN groups g ON u.group_id = g.id
		WHERE u.email = $1
	`, email).Scan(&user.ID, &user.Name, &user.PasswordHash, &user.Email, &user.GroupID, &user.Group, &user.Role, &user.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.UserDetails{}, domain.ErrNotFound
		}
		return domain.UserDetails{}, fmt.Errorf("query get user: %w", err)
	}
	return user, nil
}
