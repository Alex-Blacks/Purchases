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
