package storage

import (
	"context"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	domain.Querier
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
