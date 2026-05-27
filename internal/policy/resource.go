package policy

type ResourceOwner interface {
	OwnerID() int
}

func CanAccess(actor Actor, resource ResourceOwner) error {
	if actor.HasRole(RoleAdmin) {
		return nil
	}
	if actor.UserID == resource.OwnerID() {
		return nil
	}
	return ErrForbidden
}

func CanList(actor Actor) error {
	if actor.HasRole(RoleAdmin) {
		return nil
	}
	return ErrForbidden
}
