package policy

import "github.com/Alex-Blacks/Purchases/internal/domain"

type ResourceGroup interface {
	GetGroupID() int
}

// Разрешает админу, и участникам группы, просматривать материалы группы
func CanGroupAccessForReading(actor Actor, resource ResourceGroup) error {
	if actor.HasRole(RoleAdmin) {
		return nil
	}
	if resource.GetGroupID() == CommonGroupID {
		return nil
	}
	if actor.GroupID == resource.GetGroupID() {
		return nil
	}
	return ErrForbidden
}

// Разрешает админу, и участникам группы, изменять материалы группы
func CanGroupAccessForModify(actor Actor, resource ResourceGroup) error {
	if actor.HasRole(RoleAdmin) {
		return nil
	}
	if actor.GroupID == resource.GetGroupID() {
		return nil
	}
	return ErrForbidden
}

// Разрешает просматривать пользователя админу, и самому пользователю или семье
func CanReadUser(actor Actor, target domain.UserDetails) error {
	if actor.HasRole(RoleAdmin) {
		return nil
	}
	if actor.UserID == target.ID || actor.GroupID == target.GroupID {
		return nil
	}
	return ErrForbidden
}

// Разрешает изменять пользователя админу, и самому пользователю
func CanUpdateUser(actor Actor, target domain.UserDetails) error {
	if actor.HasRole(RoleAdmin) {
		return nil
	}
	if actor.UserID == target.ID {
		return nil
	}
	return ErrForbidden
}

// Разрешает удалять пользователя админу, пользователю разрешено удалять себя, если он не админ группы
func CanDeleteUser(actor Actor, target domain.UserDetails, isGroupAdmin bool) error {
	if actor.HasRole(RoleAdmin) {
		return nil
	}
	if actor.UserID == target.ID && !isGroupAdmin {
		return nil
	}
	return ErrForbidden
}

// Проверяет разрешение делать действия с группой. Либо админу, либо пользователю из запрашиваемой группы
func IsAccessReadGroup(actor Actor, targetGroupID int) bool {
	if actor.HasRole(RoleAdmin) {
		return true
	}
	return actor.GroupID == targetGroupID
}

// Проверяет разрешение делать действия с группой. Либо админу, либо админу группы
func IsAccessWriteGroup(actor Actor, isGroupAdmin bool) bool {
	if actor.HasRole(RoleAdmin) {
		return true
	}
	return isGroupAdmin
}
