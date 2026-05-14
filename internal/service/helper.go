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
		return fmt.Errorf("Error begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				err = fmt.Errorf("tx err: %v, rollback err: %w", err, rbErr)
			}
			return
		}

		if cmErr := tx.Commit(ctx); cmErr != nil {
			err = fmt.Errorf("commit err: %w", cmErr)
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
