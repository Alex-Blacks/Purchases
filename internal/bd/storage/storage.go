package storage

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5"
)

type Storage struct {
	mu sync.RWMutex
	db *pgx.Conn
}

func NewStorage(conn *pgx.Conn) *Storage {
	return &Storage{
		db: conn,
	}
}

func (s *Storage) Create(ctx context.Context, name string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		s.mu.Lock()
		defer s.mu.Unlock()

		return nil
	}
}
