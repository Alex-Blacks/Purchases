package policy

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type Actor struct {
	UserID int
	Role   Role
}

func (a *Actor) HasRole(role Role) bool {
	if a.Role == role {
		return true
	}
	return false
}

func ToActor(userID int, role Role) Actor {
	return Actor{
		UserID: userID,
		Role:   role,
	}
}
