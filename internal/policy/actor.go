package policy

import "github.com/Alex-Blacks/Purchases/internal/domain"

const (
	CommonGroupID int = 1
)

type Actor struct {
	UserID  int
	GroupID int
	Role    domain.UserRole
}

func (a *Actor) HasRole(role domain.UserRole) bool {
	if a.Role == role {
		return true
	}
	return false
}

func ToActor(userID int, groupID int, role domain.UserRole) Actor {
	return Actor{
		UserID:  userID,
		GroupID: groupID,
		Role:    role,
	}
}
