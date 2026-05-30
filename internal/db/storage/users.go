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

func (u *UserRepo) CreateUser(ctx context.Context, q domain.Querier, name, password_hash, email, role, status string) (domain.User, error) {
	var user domain.User
	if err := q.QueryRow(ctx, `
		INSERT INTO users(name,password_hash,email,role, status) 
		VALUES ($1,$2,$3,$4,$5) 
		RETURNING id,name,password_hash,email,role, status
	`, name, password_hash, email, role, status).Scan(&user.ID, &user.Name, &user.PasswordHash, &user.Email, &user.Role, &user.Status); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return user, domain.ErrAlreadyExists
		}
		return user, fmt.Errorf("query create user: %w", err)
	}
	return user, nil
}
func (u *UserRepo) GetUserByID(ctx context.Context, q domain.Querier, userID int) (domain.User, error) {
	var user domain.User
	if err := q.QueryRow(ctx, `
		SELECT id,name,password_hash,email,role, status 
		FROM users 
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Name, &user.PasswordHash, &user.Email, &user.Role, &user.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, domain.ErrNotFound
		}
		return user, fmt.Errorf("query get user: %w", err)
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
	rows, err := q.Query(ctx, `SELECT id,name,password_hash,email,role, status FROM users`)
	if err != nil {
		return nil, fmt.Errorf("query list users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User

		if err := rows.Scan(&user.ID, &user.Name, &user.PasswordHash, &user.Email, &user.Role, &user.Status); err != nil {
			return nil, fmt.Errorf("scan list users: %w", err)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration failed: %w", err)
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
		return user, domain.ErrNoFieldsToUpdate
	}
	if err := q.QueryRow(ctx, `
		UPDATE users
		SET `+set+`
		WHERE id = $1
		RETURNING id,name,password_hash,email,role, status
	`, args...).Scan(&user.ID, &user.Name, &user.PasswordHash, &user.Email, &user.Role, &user.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, domain.ErrNotFound
		}
		return user, fmt.Errorf("query get user: %w", err)
	}
	return user, nil
}

func (u *UserRepo) GetUserByEmail(ctx context.Context, q domain.Querier, email string) (domain.User, error) {
	var user domain.User
	if err := q.QueryRow(ctx, `
		SELECT id,name,password_hash,email,role, status 
		FROM users 
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Name, &user.PasswordHash, &user.Email, &user.Role, &user.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, domain.ErrNotFound
		}
		return user, fmt.Errorf("query get user: %w", err)
	}
	return user, nil
}
