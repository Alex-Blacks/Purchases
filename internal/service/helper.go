package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

func (s *Service) WithTx(ctx context.Context, fn func(q domain.Queryer) error) error {
	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("Error begin tx: %w", err)
	}

	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("Error commit tx: %w", err)
	}

	return nil
}
