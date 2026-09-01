package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/wneessen/go-mail"
)

func sendWithRetryAndRateLimit(ctx context.Context, logger *slog.Logger, msg *mail.Msg, client *mail.Client, maxRetries int, delay time.Duration) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = client.DialAndSend(msg)
			if err == nil {
				time.Sleep(delay)
				return nil
			}

			var sendErr *mail.SendError
			if errors.As(err, &sendErr) && sendErr.IsTemp() {
				logger.Info("Temporary error", "attempt", i+1, "error", err, "retry in", delay)
				time.Sleep(delay)
				continue
			}
			return err
		}
	}
	return err
}

func sendMail(ctx context.Context, logger *slog.Logger, client *mail.Client, from, to string, group string) error {
	msg := mail.NewMsg()
	if err := msg.From(from); err != nil {
		return err
	}
	if err := msg.To(to); err != nil {
		return err
	}
	subject := fmt.Sprintf("Вас пригласили в семью: %s", group)
	msg.Subject(subject)

	// Основное тело – текст
	textBody := fmt.Sprintf("Вас пригласили в семью: %s.\nПриглашение уже доступно в приложении.\nЕсли у вас нет приложения, скачайте его по ссылке: https://ya.ru", group)
	msg.SetBodyString(mail.TypeTextPlain, textBody)

	// HTML-альтернатива
	htmlBody := fmt.Sprintf("<h1>Приглашение</h1><p>Вас пригласили в семью <b>%s</b></p><p><a href='https://ya.ru'>Скачать приложение</a></p>", group)
	msg.AddAlternativeString(mail.TypeTextHTML, htmlBody)

	return sendWithRetryAndRateLimit(ctx, logger, msg, client, 3, 1*time.Second)
}

func validateFilterUser(filter domain.UserListFilter) error {
	if len(filter.GroupIDs) == 0 {
		return domain.ErrInvalidGroupID
	}
	if filter.Role != nil && (*filter.Role != domain.RoleAdmin && *filter.Role != domain.RoleUser) {
		return domain.ErrInvalidInput
	}
	if filter.Status != nil && (*filter.Status != domain.UserStatusActive && *filter.Status != domain.UserStatusBlocked) {
		return domain.ErrInvalidInput
	}
	if filter.CreatedFrom != nil && filter.CreatedFrom.After(*filter.CreatedTo) {
		return domain.ErrInvalidInput
	}
	if filter.CreatedTo != nil && filter.CreatedTo.Before(*filter.CreatedFrom) {
		return domain.ErrInvalidInput
	}
	if filter.UpdatedFrom != nil && filter.UpdatedFrom.After(*filter.UpdatedTo) {
		return domain.ErrInvalidInput
	}
	if filter.UpdatedTo != nil && filter.UpdatedTo.Before(*filter.UpdatedFrom) {
		return domain.ErrInvalidInput
	}
	if filter.Limit == 0 {
		return domain.ErrInvalidInput
	}
	if filter.Offset < 0 {
		return domain.ErrInvalidInput
	}
	return nil
}
