package policy

import "github.com/Alex-Blacks/Purchases/internal/domain"

func CanViewUser(actor domain.Actor, resource domain.Resource) bool {
	if actor.UserID != resource.UserID && actor.Role != "admin" {
		return false
	}
	return true
}
