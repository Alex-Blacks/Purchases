package policy

import "time"

type User struct {
	UserID int
}

type UserPolicy struct{}

func (p *UserPolicy) CanManage(actor Actor, resource User) error {
	if actor.Role == RoleAdmin {
		return nil
	}
	if actor.UserID != resource.UserID {
		return ErrForbidden
	}
	return nil
}

func (p *UserPolicy) CanList(actor Actor) error {
	if actor.Role == RoleAdmin {
		return nil
	}
	return ErrForbidden
}

type OrderResource struct {
	ID         int
	UserID     int
	StoreID    int
	ItemsCount int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type OrderPolicy struct{}

func (o *OrderPolicy) CanView(actor Actor, resource OrderResource) error {
	if actor.Role == RoleAdmin {
		return nil
	}
	if actor.UserID != resource.UserID {
		return ErrForbidden
	}
	return nil
}
