package service

import (
	"context"
	"fmt"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
)

type ServiceHistory struct {
	*BaseService
	repo domain.ChangeHistoryRepository
}

func NewServiceHistory(st domain.Storage, repo domain.ChangeHistoryRepository) *ServiceHistory {
	return &ServiceHistory{
		BaseService: &BaseService{storage: st},
		repo:        repo,
	}
}

// List возвращает список всех записей из истории с фильтрацией.
func (s *ServiceHistory) List(ctx context.Context, actor policy.Actor, filter domain.HistoryListFilter) ([]domain.ChangeHistoryDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("group_id", actor.GroupID)
	logger.InfoContext(ctx, "listing histories in group")

	if err := prepareHistoryFilter(actor, &filter); err != nil {
		return nil, err
	}

	// 1. Получение списка записей группы из БД (без транзакции)
	users, err := s.repo.List(ctx, s.storage, filter)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list histories in group", "error", err)
		return nil, fmt.Errorf("list histories: %w", err)
	}

	logger.InfoContext(ctx, "histories in group listed successfully", "count", len(users))
	return users, nil
}

// Count возвращает количество всех записей из истории с фильтрацией.
func (s *ServiceHistory) Count(ctx context.Context, actor policy.Actor, filter domain.HistoryListFilter) (int, error) {
	logger := logging.LoggerFromContext(ctx).With("group_id", actor.GroupID)
	logger.InfoContext(ctx, "counting histories in group")

	if err := prepareHistoryFilter(actor, &filter); err != nil {
		return 0, err
	}

	// 1. Получение количества записей группы из БД (без транзакции)
	count, err := s.repo.Count(ctx, s.storage, filter)
	if err != nil {
		logger.ErrorContext(ctx, "failed to count histories in group", "error", err)
		return 0, fmt.Errorf("count histories: %w", err)
	}

	logger.InfoContext(ctx, "histories in group counted successfully", "count", count)
	return count, nil
}
