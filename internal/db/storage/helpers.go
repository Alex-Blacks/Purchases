package storage

import (
	"fmt"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

func buildUserWhere(filter domain.UserListFilter) (string, []any, int) {
	args := []any{}
	setPath := []string{}
	argPos := 1

	if len(filter.GroupIDs) > 0 {
		setPath = append(setPath, fmt.Sprintf("u.group_id = ANY($%d::int[])", argPos))
		args = append(args, filter.GroupIDs)
		argPos++
	}

	if filter.Role != nil {
		setPath = append(setPath, fmt.Sprintf("u.role = $%d", argPos))
		args = append(args, filter.Role)
		argPos++
	}

	if filter.Status != nil {
		setPath = append(setPath, fmt.Sprintf("u.status = $%d", argPos))
		args = append(args, filter.Status)
		argPos++
	}

	if filter.CreatedFrom != nil {
		setPath = append(setPath, fmt.Sprintf("u.created_at >= $%d", argPos))
		args = append(args, filter.CreatedFrom)
		argPos++
	}

	if filter.CreatedTo != nil {
		setPath = append(setPath, fmt.Sprintf("u.created_at < $%d", argPos))
		args = append(args, filter.CreatedTo)
		argPos++
	}

	if filter.UpdatedFrom != nil {
		setPath = append(setPath, fmt.Sprintf("u.updated_at >= $%d", argPos))
		args = append(args, filter.UpdatedFrom)
		argPos++
	}

	if filter.UpdatedTo != nil {
		setPath = append(setPath, fmt.Sprintf("u.updated_at < $%d", argPos))
		args = append(args, filter.UpdatedTo)
		argPos++
	}

	if len(setPath) == 0 {
		return "", args, 0
	}

	return " WHERE " + strings.Join(setPath, " AND "), args, argPos
}

func buildHistoryWhere(filter domain.HistoryListFilter) (string, []any, int) {
	args := []any{}
	setPath := []string{}
	argPos := 1

	if len(filter.GroupIDs) > 0 {
		setPath = append(setPath, fmt.Sprintf("h.group_id = ANY($%d::int[])", argPos))
		args = append(args, filter.GroupIDs)
		argPos++
	}

	if filter.EntityType != nil {
		setPath = append(setPath, fmt.Sprintf("h.entity_type = $%d", argPos))
		args = append(args, filter.EntityType)
		argPos++
	}

	if filter.EntityID != nil {
		setPath = append(setPath, fmt.Sprintf("h.entity_id = $%d", argPos))
		args = append(args, filter.EntityID)
		argPos++
	}

	if filter.Action != nil {
		setPath = append(setPath, fmt.Sprintf("h.action = $%d", argPos))
		args = append(args, filter.Action)
		argPos++
	}

	if filter.From != nil {
		setPath = append(setPath, fmt.Sprintf("h.created_at >= $%d", argPos))
		args = append(args, filter.From)
		argPos++
	}

	if filter.To != nil {
		setPath = append(setPath, fmt.Sprintf("h.created_at < $%d", argPos))
		args = append(args, filter.To)
		argPos++
	}

	if len(setPath) == 0 {
		return "", args, 0
	}

	return " WHERE " + strings.Join(setPath, " AND "), args, argPos
}

func buildUnitWhere(filter domain.UnitListFilter) (string, []any, int) {
	args := []any{}
	setPath := []string{}
	argPos := 1

	if len(filter.GroupIDs) > 0 {
		setPath = append(setPath, fmt.Sprintf("u.group_id = ANY($%d::int[])", argPos))
		args = append(args, filter.GroupIDs)
		argPos++
	}

	if filter.Name != nil {
		setPath = append(setPath, fmt.Sprintf("u.name = $%d", argPos))
		args = append(args, filter.Name)
		argPos++
	}

	if filter.ShortName != nil {
		setPath = append(setPath, fmt.Sprintf("u.short_name = $%d", argPos))
		args = append(args, filter.ShortName)
		argPos++
	}

	if len(setPath) == 0 {
		return "", args, 0
	}

	return " WHERE " + strings.Join(setPath, " AND "), args, argPos
}

func buildStoreWhere(filter domain.StoreListFilter) (string, []any, int) {
	args := []any{}
	setPath := []string{}
	argPos := 1

	if len(filter.GroupIDs) > 0 {
		setPath = append(setPath, fmt.Sprintf("s.group_id = ANY($%d::int[])", argPos))
		args = append(args, filter.GroupIDs)
		argPos++
	}

	if filter.Name != nil {
		setPath = append(setPath, fmt.Sprintf("s.name = $%d", argPos))
		args = append(args, filter.Name)
		argPos++
	}

	if len(setPath) == 0 {
		return "", args, 0
	}

	return " WHERE " + strings.Join(setPath, " AND "), args, argPos
}

func buildProductWhere(filter domain.ProductListFilter) (string, []any, int) {
	args := []any{}
	setPath := []string{}
	argPos := 1

	if len(filter.GroupIDs) > 0 {
		setPath = append(setPath, fmt.Sprintf("p.group_id = ANY($%d::int[])", argPos))
		args = append(args, filter.GroupIDs)
		argPos++
	}

	if filter.Title != nil {
		setPath = append(setPath, fmt.Sprintf("p.title = $%d", argPos))
		args = append(args, filter.Title)
		argPos++
	}

	if len(setPath) == 0 {
		return "", args, 0
	}

	return " WHERE " + strings.Join(setPath, " AND "), args, argPos
}

func buildProductAliasWhere(filter domain.ProductAliasListFilter) (string, []any, int) {
	args := []any{}
	setPath := []string{}
	argPos := 1

	if len(filter.GroupIDs) > 0 {
		setPath = append(setPath, fmt.Sprintf("pa.group_id = ANY($%d::int[])", argPos))
		args = append(args, filter.GroupIDs)
		argPos++
	}

	if filter.ProductID > 0 {
		setPath = append(setPath, fmt.Sprintf("pa.product_id = $%d", argPos))
		args = append(args, filter.ProductID)
		argPos++
	}

	if filter.Alias != nil {
		setPath = append(setPath, fmt.Sprintf("pa.alias = $%d", argPos))
		args = append(args, filter.Alias)
		argPos++
	}

	if len(setPath) == 0 {
		return "", args, 0
	}

	return " WHERE " + strings.Join(setPath, " AND "), args, argPos
}

func buildOrderWhere(filter domain.OrderListFilter) (string, []any, int) {
	args := []any{}
	setPath := []string{}
	argPos := 1

	if len(filter.GroupIDs) > 0 {
		setPath = append(setPath, fmt.Sprintf("o.group_id = ANY($%d::int[])", argPos))
		args = append(args, filter.GroupIDs)
		argPos++
	}

	if filter.UserID != nil {
		setPath = append(setPath, fmt.Sprintf("o.user_id = $%d", argPos))
		args = append(args, filter.UserID)
		argPos++
	}
	if filter.StoreID != nil {
		setPath = append(setPath, fmt.Sprintf("o.store_id = $%d", argPos))
		args = append(args, filter.StoreID)
		argPos++
	}
	if filter.CreatedFrom != nil {
		setPath = append(setPath, fmt.Sprintf("o.created_at >= $%d", argPos))
		args = append(args, filter.CreatedFrom)
		argPos++
	}

	if filter.CreatedTo != nil {
		setPath = append(setPath, fmt.Sprintf("o.created_at < $%d", argPos))
		args = append(args, filter.CreatedTo)
		argPos++
	}

	if filter.UpdatedFrom != nil {
		setPath = append(setPath, fmt.Sprintf("o.updated_at >= $%d", argPos))
		args = append(args, filter.UpdatedFrom)
		argPos++
	}

	if filter.UpdatedTo != nil {
		setPath = append(setPath, fmt.Sprintf("o.updated_at < $%d", argPos))
		args = append(args, filter.UpdatedTo)
		argPos++
	}

	if len(setPath) == 0 {
		return "", args, 0
	}

	return " WHERE " + strings.Join(setPath, " AND "), args, argPos
}

func buildOrderItemWhere(filter domain.OrderItemListFilter) (string, []any, int) {
	args := []any{}
	setPath := []string{}
	argPos := 1

	if len(filter.GroupIDs) > 0 {
		setPath = append(setPath, fmt.Sprintf("oi.group_id = ANY($%d::int[])", argPos))
		args = append(args, filter.GroupIDs)
		argPos++
	}

	if filter.OrderID > 0 {
		setPath = append(setPath, fmt.Sprintf("oi.order_id = $%d", argPos))
		args = append(args, filter.OrderID)
		argPos++
	}
	if filter.ProductID != nil {
		setPath = append(setPath, fmt.Sprintf("oi.product_id = $%d", argPos))
		args = append(args, filter.ProductID)
		argPos++
	}
	if filter.UnitID != nil {
		setPath = append(setPath, fmt.Sprintf("oi.unit_id = $%d", argPos))
		args = append(args, filter.UnitID)
		argPos++
	}

	if filter.Quantity != nil {
		setPath = append(setPath, fmt.Sprintf("oi.quantity = $%d", argPos))
		args = append(args, filter.Quantity)
		argPos++
	}

	if len(setPath) == 0 {
		return "", args, 0
	}

	return " WHERE " + strings.Join(setPath, " AND "), args, argPos
}

func buildInviteWhere(filter domain.InviteListFilter) (string, []any, int) {
	args := []any{}
	setPath := []string{}
	argPos := 1

	if len(filter.GroupIDs) > 0 {
		setPath = append(setPath, fmt.Sprintf("i.group_id = ANY($%d::int[])", argPos))
		args = append(args, filter.GroupIDs)
		argPos++
	}

	if filter.InviterUserID != nil {
		setPath = append(setPath, fmt.Sprintf("i.inviter_user_id = $%d", argPos))
		args = append(args, filter.InviterUserID)
		argPos++
	}

	if filter.InviteeEmail != nil {
		setPath = append(setPath, fmt.Sprintf("i.invitee_email = $%d", argPos))
		args = append(args, filter.InviteeEmail)
		argPos++
	}

	if filter.Status != nil {
		setPath = append(setPath, fmt.Sprintf("i.status = $%d", argPos))
		args = append(args, filter.Status)
		argPos++
	}

	if filter.Token != nil {
		setPath = append(setPath, fmt.Sprintf("i.token = $%d", argPos))
		args = append(args, filter.Token)
		argPos++
	}

	if filter.CreatedFrom != nil {
		setPath = append(setPath, fmt.Sprintf("i.created_at >= $%d", argPos))
		args = append(args, filter.CreatedFrom)
		argPos++
	}

	if filter.CreatedTo != nil {
		setPath = append(setPath, fmt.Sprintf("i.created_at < $%d", argPos))
		args = append(args, filter.CreatedTo)
		argPos++
	}

	if filter.ExpiresFrom != nil {
		setPath = append(setPath, fmt.Sprintf("i.expires_at >= $%d", argPos))
		args = append(args, filter.ExpiresFrom)
		argPos++
	}

	if filter.ExpiresTo != nil {
		setPath = append(setPath, fmt.Sprintf("i.expires_at < $%d", argPos))
		args = append(args, filter.ExpiresTo)
		argPos++
	}

	if len(setPath) == 0 {
		return "", args, 0
	}

	return " WHERE " + strings.Join(setPath, " AND "), args, argPos
}

func buildGroupWhere(filter domain.GroupListFilter) (string, []any, int) {
	args := []any{}
	setPath := []string{}
	argPos := 1

	if filter.Name != nil {
		setPath = append(setPath, fmt.Sprintf("g.name = $%d", argPos))
		args = append(args, filter.Name)
		argPos++
	}

	if filter.AdminUserID != nil {
		setPath = append(setPath, fmt.Sprintf("g.admin_user_id = $%d", argPos))
		args = append(args, filter.AdminUserID)
		argPos++
	}

	if len(setPath) == 0 {
		return "", args, 0
	}

	return " WHERE " + strings.Join(setPath, " AND "), args, argPos
}
