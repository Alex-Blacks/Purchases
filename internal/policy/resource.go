package policy

import (
	"context"
)

func (a *Actor) HasRole(role Role) bool {
	for _, r := range a.Roles {
		if r == role {
			return true
		}

	}
	return false
}

type ResurceOwner interface {
	OwnerID() int
}

type User struct {
	UserID int
}

func (u User) OwnerID() int { return u.UserID }

type Order struct {
	ID     int
	UserID int
}

func (o Order) OwnerID() int { return o.UserID }

type OwnerPolicy struct{}

func (OwnerPolicy) CanAccess(ctx context.Context, actor Actor, resource ResurceOwner) error {
	if actor.HasRole(RoleAdmin) {
		return nil
	}
	if actor.UserID == resource.OwnerID() {
		return nil
	}
	return ErrForbidden
}

type UserPolicy struct {
	OwnerPolicy
}

func (UserPolicy) CanList(ctx context.Context, actor Actor) error {
	if actor.HasRole(RoleAdmin) {
		return nil
	}
	return ErrForbidden
}

func (u UserPolicy) CanManage(ctx context.Context, actor Actor, resource User) error {
	return u.CanAccess(ctx, actor, resource)

}

type OrderPolicy struct {
	OwnerPolicy
}

func (OrderPolicy) CanList(ctx context.Context, actor Actor) error {
	if actor.HasRole(RoleAdmin) {
		return nil
	}
	return ErrForbidden
}

func (o OrderPolicy) CanManage(ctx context.Context, actor Actor, resource Order) error {
	return o.CanAccess(ctx, actor, resource)

}
