package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/logging"
	"github.com/Alex-Blacks/Purchases/internal/policy"
	"github.com/google/uuid"
)

type ServiceInvite struct {
	storage domain.Storage
	invite  domain.InviteRepository
	group   domain.GroupRepository
	user    domain.UserRepository
}

func NewServiceInvite(st domain.Storage, invite domain.InviteRepository, group domain.GroupRepository, user domain.UserRepository) *ServiceInvite {
	return &ServiceInvite{
		storage: st,
		invite:  invite,
		group:   group,
		user:    user,
	}
}

// WithTx выполняет функцию в транзакции.
func (s *ServiceInvite) WithTx(ctx context.Context, fn func(q domain.Querier) error) (err error) {
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

// getValidPendingInvite возвращает приглашение, если оно действительно и ожидает подтверждения.
// Возвращает ошибку, если токен недействителен, срок истёк, статус не pending или email не совпадает.
func (s *ServiceInvite) getValidPendingInvite(ctx context.Context, q domain.Querier, logger *slog.Logger, actor policy.Actor, token string) (domain.InviteDetails, error) {
	// 1. Получение приглашения по токену
	existing, err := s.invite.GetInviteByToken(ctx, q, token)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.WarnContext(ctx, "invalid token")
			return domain.InviteDetails{}, policy.ErrForbidden
		}
		logger.ErrorContext(ctx, "failed to get invite by token", "error", err)
		return domain.InviteDetails{}, fmt.Errorf("get invite by token: %w", err)
	}

	// 2. Проверка статус
	switch existing.Status {
	case "accepted":
		logger.InfoContext(ctx, "invite already accepted", "invite_id", existing.ID)
		return domain.InviteDetails{}, domain.ErrUserAlreadyInGroup
	case "rejected":
		logger.InfoContext(ctx, "invite rejected", "invite_id", existing.ID)
		return domain.InviteDetails{}, domain.ErrInviteRejected
	case "pending":
		// 3. Проверка срока действия
		if existing.ExpiresAt.Before(time.Now()) {
			logger.InfoContext(ctx, "invite expired", "invite_id", existing.ID)
			return domain.InviteDetails{}, domain.ErrInviteExpired
		}

		// 4. Проверка, что пользователь совпадает с invitee_email
		user, err := s.user.GetUserByID(ctx, q, actor.UserID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to get user", "error", err)
			return domain.InviteDetails{}, fmt.Errorf("get user: %w", err)
		}
		if user.Email != existing.InviteeEmail {
			logger.WarnContext(ctx, "user email mismatch", "user_email", user.Email, "invitee_email", existing.InviteeEmail)
			return domain.InviteDetails{}, policy.ErrForbidden
		}
	default:
		logger.WarnContext(ctx, "unknown invite status", "status", existing.Status)
		return domain.InviteDetails{}, fmt.Errorf("unknown invite status: %s", existing.Status)
	}
	return existing, nil
}

// CreateInvite создаёт приглашение для пользователя с email inviteeEmail в группу actor.GroupID.
func (s *ServiceInvite) CreateInvite(ctx context.Context, actor policy.Actor, inviteeEmail string) (domain.InviteDetails, error) {
	logger := logging.LoggerFromContext(ctx).With("invitee_email", inviteeEmail)
	logger.InfoContext(ctx, "creating new invite")

	var result domain.InviteDetails
	if err := s.WithTx(ctx, func(q domain.Querier) error {

		// 1. Проверка прав: только администратор группы может приглашать
		if !s.group.CheckGroupAdmin(ctx, q, actor.GroupID, actor.UserID) {
			logger.WarnContext(ctx, "user is not group admin")
			return policy.ErrForbidden
		}

		// 2. Поиск пользователя по email
		invitee, err := s.user.GetUserByEmail(ctx, q, inviteeEmail)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			logger.ErrorContext(ctx, "failed to get user by email", "error", err)
			return fmt.Errorf("get user by email: %w", err)
		}

		// 3. Если пользователь существует — проверяем, что он не состоит в группе
		if invitee.ID != 0 {
			// 3a. Нельзя пригласить самого себя
			if invitee.ID == actor.UserID {
				logger.InfoContext(ctx, "attempt to invite yourself")
				return domain.ErrSelfInvite
			}
			// 3b. Нельзя пригласить того, кто уже в группе
			if invitee.GroupID == actor.GroupID {
				logger.InfoContext(ctx, "user already in group", "user_id", invitee.ID)
				return domain.ErrUserAlreadyInGroup
			}
		}

		// 4. Проверяем, есть ли уже активное (pending) приглашение
		existing, err := s.invite.GetInviteByEmail(ctx, q, actor.GroupID, inviteeEmail)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			logger.ErrorContext(ctx, "failed to get invite by email", "error", err)
			return fmt.Errorf("get invite by email: %w", err)
		}
		if existing.ID != 0 {
			switch existing.Status {
			// 4a. Если приглашение существует и оно в статусе pending — ошибка
			case "pending":
				logger.InfoContext(ctx, "pending invite already exists", "invite_id", existing.ID)
				return domain.ErrInviteAlreadyPending
			// 4b. Если приглашение существует и оно в статусе accepted — ошибка
			case "accepted":
				logger.InfoContext(ctx, "pending invite already accepted", "invite_id", existing.ID)
				return domain.ErrUserAlreadyInGroup
			case "rejected":
				// 4c. Если приглашение отклонено и срок истёк — можно создать новое
				if existing.ExpiresAt.Before(time.Now()) {
					logger.InfoContext(ctx, "rejected invite expired, removing and creating new one", "invite_id", existing.ID)
					if err := s.invite.DeleteInviteByID(ctx, q, existing.ID, existing.GroupID); err != nil {
						logger.ErrorContext(ctx, "failed to delete expired invite", "error", err)
						return fmt.Errorf("delete expired invite: %w", err)
					}
				} else {
					// 4e. Если приглашение отклонено, но срок ещё не истёк — пока запрещаем создавать новое
					logger.InfoContext(ctx, "rejected invite still valid", "invite_id", existing.ID)
					return domain.ErrInviteRejectedStillValid
				}
			default:
				logger.WarnContext(ctx, "unknown invite status", "status", existing.Status)
				return fmt.Errorf("unknown invite status: %s", existing.Status)
			}
		}

		// 5. Генерация токена
		token := uuid.New().String()

		// 6. Создание приглашения
		created, err := s.invite.CreateInvite(ctx, q, actor.GroupID, actor.UserID, inviteeEmail, token)
		if err != nil {
			logger.ErrorContext(ctx, "failed to create invite", "error", err)
			return fmt.Errorf("create invite: %w", err)
		}

		result = created
		return nil

	}); err != nil {
		return domain.InviteDetails{}, fmt.Errorf("create invite: %w", err)
	}

	logger.InfoContext(ctx, "invite created successfully", "invite_id", result.ID)
	return result, nil
}

func (s *ServiceInvite) AcceptInvite(ctx context.Context, actor policy.Actor, token string) error {
	logger := logging.LoggerFromContext(ctx).With("token", token)
	logger.InfoContext(ctx, "accepting invite")

	return s.WithTx(ctx, func(q domain.Querier) error {

		// 1. Проверка статуса, срока и email
		existing, err := s.getValidPendingInvite(ctx, q, logger, actor, token)
		if err != nil {
			return err
		}

		// 2. Проверка что пользователь уже не в группе
		if actor.GroupID == existing.GroupID {
			return domain.ErrUserAlreadyInGroup
		}

		// 2. Обновить статус приглашения на 'accepted'
		status := domain.StatusAccepted
		if _, err := s.invite.UpdateInviteByID(ctx, q, existing.ID, existing.GroupID, domain.InviteUpdate{Status: &status}); err != nil {
			logger.ErrorContext(ctx, "failed to update invite status", "error", err)
			return fmt.Errorf("update invite: %w", err)
		}

		// 3. Добавить пользователя в группу
		if _, err := s.user.UpdateUserByID(ctx, q, actor.UserID, domain.UserUpdate{GroupID: &existing.GroupID}); err != nil {
			logger.ErrorContext(ctx, "failed to update user group", "error", err)
			return fmt.Errorf("update user: %w", err)
		}
		logger.InfoContext(ctx, "invite accepted successfully", "invite_id", existing.ID)
		return nil

	})
}

func (s *ServiceInvite) RejectInvite(ctx context.Context, actor policy.Actor, token string) error {
	logger := logging.LoggerFromContext(ctx).With("token", token)
	logger.InfoContext(ctx, "rejecting invite")

	return s.WithTx(ctx, func(q domain.Querier) error {

		// 1. Проверка статуса, срока и email
		existing, err := s.getValidPendingInvite(ctx, q, logger, actor, token)
		if err != nil {
			return err
		}

		// 2. Обновить статус приглашения на 'rejected'
		status := domain.StatusRejected
		if _, err := s.invite.UpdateInviteByID(ctx, q, existing.ID, existing.GroupID, domain.InviteUpdate{Status: &status}); err != nil {
			logger.ErrorContext(ctx, "failed to update invite status", "error", err)
			return fmt.Errorf("update invite: %w", err)
		}

		logger.InfoContext(ctx, "invite rejected successfully", "invite_id", existing.ID)
		return nil
	})
}

func (s *ServiceInvite) ListInvites(ctx context.Context, actor policy.Actor) ([]domain.InviteDetails, error) {
	logger := logging.LoggerFromContext(ctx)
	logger.InfoContext(ctx, "listing invites")

	// 1. Проверка прав: только администратор группы может смотреть список приглашений
	if !s.group.CheckGroupAdmin(ctx, s.storage, actor.GroupID, actor.UserID) {
		logger.WarnContext(ctx, "user is not group admin")
		return []domain.InviteDetails{}, policy.ErrForbidden
	}

	// 2. Получение списка приглашений
	result, err := s.invite.ListInvites(ctx, s.storage, actor.GroupID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list invites", "error", err)
		return []domain.InviteDetails{}, fmt.Errorf("list invites: %w", err)
	}

	logger.InfoContext(ctx, "invites list successfully", "count", len(result))
	return result, nil
}
