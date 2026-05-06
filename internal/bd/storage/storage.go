package storage

import (
	"context"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{
		pool: pool,
	}
}

func (s *Storage) BeginTx(ctx context.Context) (domain.Tx, error) {
	return s.pool.Begin(ctx)
}

func (s *Storage) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return s.pool.Exec(ctx, sql, args...)
}

func (s *Storage) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return s.pool.Query(ctx, sql, args...)
}

func (s *Storage) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return s.pool.QueryRow(ctx, sql, args...)
}
