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

func CanReadUser(actor Actor, target domain.UserDetails) error {
	if actor.HasRole(RoleAdmin) {
		return nil
	}
	if actor.UserID == target.ID && actor.GroupID == target.GroupID {
		return nil
	}
	return ErrForbidden
}

func CanUpdateUser(actor Actor, target domain.UserDetails) error {
	if actor.HasRole(RoleAdmin) {
		return nil
	}
	if actor.UserID == target.ID {
		return nil
	}
	return ErrForbidden
}

func CanDeleteUser(actor Actor, target domain.UserDetails, isGroupAdmin bool) error {
	if actor.HasRole(RoleAdmin) {
		return nil
	}
	if actor.UserID == target.ID || !isGroupAdmin {
		return nil
	}
	return ErrForbidden
}
