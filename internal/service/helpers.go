package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/policy"
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

func contains(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

func prepareCommonFilter(actor policy.Actor, groupIDs []int, limit, offset int) ([]int, error) {
	if limit <= 0 || offset < 0 {
		return nil, domain.ErrInvalidInput
	}
	if !actor.HasRole(domain.RoleAdmin) {
		allowed := []int{actor.GroupID, policy.CommonGroupID}
		if len(groupIDs) == 0 {
			return allowed, nil
		}
		for _, gid := range groupIDs {
			if !contains(allowed, gid) {
				return nil, policy.ErrForbidden
			}
		}
		return groupIDs, nil
	}
	// Администратор: проверяем, что все ID > 0
	for _, gid := range groupIDs {
		if gid < 1 {
			return nil, domain.ErrInvalidGroupID
		}
	}
	return groupIDs, nil
}

// prepareUserFilter валидирует и подготавливает фильтр для User.
// Модифицирует filter.GroupIDs в зависимости от роли актора.
func prepareUserFilter(actor policy.Actor, filter *domain.UserListFilter) error {
	// 1. Валидация фильтра (если передано)
	if filter.Role != nil && (*filter.Role != domain.RoleAdmin && *filter.Role != domain.RoleUser) {
		return domain.ErrInvalidInput
	}
	if filter.Status != nil && (*filter.Status != domain.UserStatusActive && *filter.Status != domain.UserStatusBlocked) {
		return domain.ErrInvalidInput
	}
	if filter.CreatedFrom != nil && filter.CreatedTo != nil && filter.CreatedFrom.After(*filter.CreatedTo) {
		return domain.ErrInvalidInput
	}
	if filter.CreatedTo != nil && filter.CreatedFrom != nil && filter.CreatedTo.Before(*filter.CreatedFrom) {
		return domain.ErrInvalidInput
	}
	if filter.UpdatedFrom != nil && filter.UpdatedTo != nil && filter.UpdatedFrom.After(*filter.UpdatedTo) {
		return domain.ErrInvalidInput
	}
	if filter.UpdatedTo != nil && filter.UpdatedFrom != nil && filter.UpdatedTo.Before(*filter.UpdatedFrom) {
		return domain.ErrInvalidInput
	}
	var err error
	filter.GroupIDs, err = prepareCommonFilter(actor, filter.GroupIDs, filter.Limit, filter.Offset)
	return err
}

// prepareUnitFilter валидирует и подготавливает фильтр для Unit.
// Модифицирует filter.GroupIDs в зависимости от роли актора.
func prepareUnitFilter(actor policy.Actor, filter *domain.UnitListFilter) error {
	// 1. Валидация фильтра по имени (если передано)
	if filter.Name != nil && strings.TrimSpace(*filter.Name) == "" {
		return domain.ErrEmptyName
	}
	if filter.ShortName != nil && strings.TrimSpace(*filter.ShortName) == "" {
		return domain.ErrEmptyName
	}

	var err error
	filter.GroupIDs, err = prepareCommonFilter(actor, filter.GroupIDs, filter.Limit, filter.Offset)
	return err
}

// prepareStoreFilter валидирует и подготавливает фильтр для Store.
// Модифицирует filter.GroupIDs в зависимости от роли актора.
func prepareStoreFilter(actor policy.Actor, filter *domain.StoreListFilter) error {
	// 1. Валидация фильтра по имени (если передано)
	if filter.Name != nil && strings.TrimSpace(*filter.Name) == "" {
		return domain.ErrEmptyName
	}

	var err error
	filter.GroupIDs, err = prepareCommonFilter(actor, filter.GroupIDs, filter.Limit, filter.Offset)
	return err
}

// prepareProductFilter валидирует и подготавливает фильтр для Product.
// Модифицирует filter.GroupIDs в зависимости от роли актора.
func prepareProductFilter(actor policy.Actor, filter *domain.ProductListFilter) error {
	// 1. Валидация фильтра по имени (если передано)
	if filter.Title != nil && strings.TrimSpace(*filter.Title) == "" {
		return domain.ErrEmptyName
	}

	var err error
	filter.GroupIDs, err = prepareCommonFilter(actor, filter.GroupIDs, filter.Limit, filter.Offset)
	return err
}

// prepareProductAliasFilter валидирует и подготавливает фильтр для ProductAlias.
// Модифицирует filter.GroupIDs в зависимости от роли актора.
func prepareProductAliasFilter(actor policy.Actor, filter *domain.ProductAliasListFilter) error {
	// 1. Валидация фильтра (если передано)
	if filter.Alias != nil && strings.TrimSpace(*filter.Alias) == "" {
		return domain.ErrEmptyName
	}

	if filter.ProductID < 1 {
		return domain.ErrInvalidInput
	}

	var err error
	filter.GroupIDs, err = prepareCommonFilter(actor, filter.GroupIDs, filter.Limit, filter.Offset)
	return err
}

// prepareOrderFilter валидирует и подготавливает фильтр для Order.
// Модифицирует filter.GroupIDs в зависимости от роли актора.
func prepareOrderFilter(actor policy.Actor, filter *domain.OrderListFilter) error {
	// 1. Валидация фильтра (если передано)
	if filter.UserID != nil && *filter.UserID < 1 {
		return domain.ErrInvalidInput
	}
	if filter.StoreID != nil && *filter.StoreID < 1 {
		return domain.ErrInvalidInput
	}
	if filter.CreatedFrom != nil && filter.CreatedTo != nil && filter.CreatedFrom.After(*filter.CreatedTo) {
		return domain.ErrInvalidInput
	}
	if filter.CreatedTo != nil && filter.CreatedFrom != nil && filter.CreatedTo.Before(*filter.CreatedFrom) {
		return domain.ErrInvalidInput
	}
	if filter.UpdatedFrom != nil && filter.UpdatedTo != nil && filter.UpdatedFrom.After(*filter.UpdatedTo) {
		return domain.ErrInvalidInput
	}
	if filter.UpdatedTo != nil && filter.UpdatedFrom != nil && filter.UpdatedTo.Before(*filter.UpdatedFrom) {
		return domain.ErrInvalidInput
	}

	var err error
	filter.GroupIDs, err = prepareCommonFilter(actor, filter.GroupIDs, filter.Limit, filter.Offset)
	return err
}

// prepareOrderItemFilter валидирует и подготавливает фильтр для OrderItem.
// Модифицирует filter.GroupIDs в зависимости от роли актора.
func prepareOrderItemFilter(actor policy.Actor, filter *domain.OrderItemListFilter) error {
	// 1. Валидация фильтра (если передано)
	if filter.OrderID < 1 {
		return domain.ErrInvalidInput
	}
	if filter.ProductID != nil && *filter.ProductID < 1 {
		return domain.ErrInvalidInput
	}
	if filter.UnitID != nil && *filter.UnitID < 1 {
		return domain.ErrInvalidInput
	}
	if filter.QuantityMin != nil && filter.QuantityMax != nil && *filter.QuantityMin > *filter.QuantityMax {
		return domain.ErrInvalidInput
	}

	var err error
	filter.GroupIDs, err = prepareCommonFilter(actor, filter.GroupIDs, filter.Limit, filter.Offset)
	return err
}

// prepareInviteFilter валидирует и подготавливает фильтр для Invite.
// Модифицирует filter.GroupIDs в зависимости от роли актора.
func prepareInviteFilter(actor policy.Actor, filter *domain.InviteListFilter) error {
	// 1. Валидация фильтра (если передано)
	if filter.InviterUserID != nil && *filter.InviterUserID < 1 {
		return domain.ErrInvalidInput
	}
	if filter.InviteeEmail != nil && strings.TrimSpace(*filter.InviteeEmail) == "" {
		return domain.ErrInvalidInput
	}
	if filter.Status != nil && *filter.Status != domain.StatusAccepted && *filter.Status != domain.StatusPending && *filter.Status != domain.StatusRejected {
		return domain.ErrInvalidInput
	}
	if filter.Token != nil && strings.TrimSpace(*filter.Token) == "" {
		return domain.ErrInvalidInput
	}
	if filter.CreatedFrom != nil && filter.CreatedTo != nil && filter.CreatedFrom.After(*filter.CreatedTo) {
		return domain.ErrInvalidInput
	}
	if filter.CreatedTo != nil && filter.CreatedFrom != nil && filter.CreatedTo.Before(*filter.CreatedFrom) {
		return domain.ErrInvalidInput
	}
	if filter.ExpiresFrom != nil && filter.ExpiresTo != nil && filter.ExpiresFrom.After(*filter.ExpiresTo) {
		return domain.ErrInvalidInput
	}
	if filter.ExpiresTo != nil && filter.ExpiresFrom != nil && filter.ExpiresTo.Before(*filter.ExpiresFrom) {
		return domain.ErrInvalidInput
	}

	var err error
	filter.GroupIDs, err = prepareCommonFilter(actor, filter.GroupIDs, filter.Limit, filter.Offset)
	return err
}

// prepareHistoryFilter валидирует и подготавливает фильтр для ChangeHistory.
// Модифицирует filter.GroupIDs в зависимости от роли актора.
func prepareHistoryFilter(actor policy.Actor, filter *domain.HistoryListFilter) error {
	// 1. Валидация фильтра (если передано)
	if filter.EntityType != nil &&
		*filter.EntityType != domain.HistoryEntityOrder &&
		*filter.EntityType != domain.HistoryEntityOrderItem &&
		*filter.EntityType != domain.HistoryEntityStore &&
		*filter.EntityType != domain.HistoryEntityUnit &&
		*filter.EntityType != domain.HistoryEntityProduct &&
		*filter.EntityType != domain.HistoryEntityProductAlias {
		return domain.ErrInvalidInput
	}

	if filter.EntityID != nil && *filter.EntityID < 1 {
		return domain.ErrInvalidInput
	}
	if filter.Action != nil &&
		*filter.Action != domain.HistoryActionCreate &&
		*filter.Action != domain.HistoryActionUpdate &&
		*filter.Action != domain.HistoryActionDelete {
		return domain.ErrInvalidInput
	}
	if filter.From != nil && filter.To != nil && filter.From.After(*filter.To) {
		return domain.ErrInvalidInput
	}
	if filter.To != nil && filter.From != nil && filter.To.Before(*filter.From) {
		return domain.ErrInvalidInput
	}

	var err error
	filter.GroupIDs, err = prepareCommonFilter(actor, filter.GroupIDs, filter.Limit, filter.Offset)
	return err
}

// validateGroupFilter валидирует фильтр для Group.
func validateGroupFilter(filter domain.GroupListFilter) error {
	// 1. Валидация фильтра (если передано)
	if filter.Name != nil && strings.TrimSpace(*filter.Name) == "" {
		return domain.ErrInvalidInput
	}
	if filter.AdminUserID != nil && *filter.AdminUserID < 1 {
		return domain.ErrInvalidInput
	}

	if filter.Limit <= 0 || filter.Offset < 0 {
		return domain.ErrInvalidInput
	}

	return nil
}
