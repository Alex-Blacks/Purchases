package policy

type ResourceGroup interface {
	GetGroupID() int
}

func CanGroupAccessForReading(actor Actor, resource ResourceGroup) error {
	if actor.HasRole(RoleAdmin) {
		return nil
	}
	if resource.GetGroupID() == commonGroupID {
		return nil
	}
	if actor.GroupID == resource.GetGroupID() {
		return nil
	}
	return ErrForbidden
}

func CanGroupAccessForModify(actor Actor, resource ResourceGroup) error {
	if actor.HasRole(RoleAdmin) {
		return nil
	}
	if actor.GroupID == resource.GetGroupID() {
		return nil
	}
	return ErrForbidden
}
