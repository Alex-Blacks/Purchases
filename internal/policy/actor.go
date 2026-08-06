package policy

type Role string

const (
	CommonGroupID int  = 1
	RoleUser      Role = "user"
	RoleAdmin     Role = "admin"
)

type Actor struct {
	UserID  int
	GroupID int
	Role    Role
}

func (a *Actor) HasRole(role Role) bool {
	if a.Role == role {
		return true
	}
	return false
}

func ToActor(userID int, groupID int, role Role) Actor {
	return Actor{
		UserID:  userID,
		GroupID: groupID,
		Role:    role,
	}
}
