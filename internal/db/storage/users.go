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

func (u *UserRepo) CreateUser(ctx context.Context, q domain.Querier, name, password_hash, email string, group_id int, role, status string) (domain.User, error) {
	var user domain.User
	if err := q.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO users(name, password_hash, email, group_id, role, status) 
			VALUES ($1,$2,$3,$4,$5,$6) 
			RETURNING id, name, password_hash, email, group_id, role, status
		)
		SELECT i.id, i.name, i.password_hash, i.email, i.group_id, g.name, i.role, i.status
		FROM inserted i
		JOIN groups g ON i.group_id = g.id
	`, name, password_hash, email, group_id, role, status).Scan(&user.ID, &user.Name, &user.PasswordHash, &user.Email, &user.GroupID, &user.Group, &user.Role, &user.Status); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.User{}, domain.ErrAlreadyExists
		}
		return domain.User{}, fmt.Errorf("query create user: %w", err)
	}
	return user, nil
}
func (u *UserRepo) GetUserByID(ctx context.Context, q domain.Querier, userID int) (domain.User, error) {
	var user domain.User
	if err := q.QueryRow(ctx, `
		SELECT u.id, u.name, u.password_hash, u.email, u.group_id, g.name, u.role, u.status 
		FROM users u
		JOIN groups g ON u.group_id = g.id
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Name, &user.PasswordHash, &user.Email, &user.GroupID, &user.Group, &user.Role, &user.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("query get user: %w", err)
	}
	return user, nil
}
func (u *UserRepo) DeleteUser(ctx context.Context, q domain.Querier, userID int) error {
	var id int
	if err := q.QueryRow(ctx, `DELETE FROM users WHERE id = $1 RETURNING id`, userID).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("query delete user: %w", err)
	}
	return nil
}
func (u *UserRepo) ListUsers(ctx context.Context, q domain.Querier) ([]domain.User, error) {
	rows, err := q.Query(ctx, `
		SELECT u.id, u.name, u.password_hash, u.email, u.group_id, g.name, u.role, u.status 
		FROM users u
		JOIN groups g ON u.group_id = g.id
	`)
	if err != nil {
		return []domain.User{}, fmt.Errorf("query list users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User

		if err := rows.Scan(&user.ID, &user.Name, &user.PasswordHash, &user.Email, &user.GroupID, &user.Group, &user.Role, &user.Status); err != nil {
			return []domain.User{}, fmt.Errorf("scan list users: %w", err)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return []domain.User{}, fmt.Errorf("iteration failed: %w", err)
	}
	return users, nil
}
func (u *UserRepo) UpdateUser(ctx context.Context, q domain.Querier, userID int, updateUser domain.UpdateUser) (domain.User, error) {
	var user domain.User
	args := []any{userID}
	setParts := []string{}
	argPos := 2
	if updateUser.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argPos))
		args = append(args, *updateUser.Name)
		argPos++
	}
	if updateUser.Password != nil {
		setParts = append(setParts, fmt.Sprintf("password_hash = $%d", argPos))
		args = append(args, *updateUser.Password)
		argPos++
	}
	if updateUser.Email != nil {
		setParts = append(setParts, fmt.Sprintf("email = $%d", argPos))
		args = append(args, *updateUser.Email)
		argPos++
	}
	if updateUser.GroupID != nil {
		setParts = append(setParts, fmt.Sprintf("group_id = $%d", argPos))
		args = append(args, *updateUser.GroupID)
		argPos++
	}
	if updateUser.Role != nil {
		setParts = append(setParts, fmt.Sprintf("role = $%d", argPos))
		args = append(args, *updateUser.Role)
		argPos++
	}
	if updateUser.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argPos))
		args = append(args, *updateUser.Status)
		argPos++
	}

	set := strings.Join(setParts, ", ")
	if strings.TrimSpace(set) == "" {
		return domain.User{}, domain.ErrNoFieldsToUpdate
	}
	if err := q.QueryRow(ctx, `
		UPDATE users u
		SET `+set+`
		FROM groups g
		WHERE u.id = $1 AND u.group_id = g.id
		RETURNING u.id, u.name, u.password_hash, u.email, u.group_id, g.name, u.role, u.status
	`, args...).Scan(&user.ID, &user.Name, &user.PasswordHash, &user.Email, &user.GroupID, &user.Group, &user.Role, &user.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("query get user: %w", err)
	}
	return user, nil
}

func (u *UserRepo) GetUserByEmail(ctx context.Context, q domain.Querier, email string) (domain.User, error) {
	var user domain.User
	if err := q.QueryRow(ctx, `
		SELECT u.id, u.name, u.password_hash, u.email, u.group_id, g.name, u.role, u.status 
		FROM users u
		JOIN groups g ON u.group_id = g.id
		WHERE u.email = $1
	`, email).Scan(&user.ID, &user.Name, &user.PasswordHash, &user.Email, &user.GroupID, &user.Group, &user.Role, &user.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("query get user: %w", err)
	}
	return user, nil
}
