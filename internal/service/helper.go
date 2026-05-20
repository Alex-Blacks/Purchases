package service

import (
	"context"
	"fmt"
	"reflect"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

func (s *Service) WithTx(ctx context.Context, fn func(q domain.Querier) error) (err error) {
	tx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = fmt.Errorf("tx err: %v, rollback err: %w", err, rollbackErr)
			}
			return
		}

		if commitErr := tx.Commit(ctx); commitErr != nil {
			err = fmt.Errorf("commit err: %w", commitErr)
		}
	}()

	err = fn(tx)
	return err
}

func hasUpdates(u domain.UpdateUser) bool {
	v := reflect.ValueOf(u)
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).IsNil() {
			return true
		}
	}
	return false
}
