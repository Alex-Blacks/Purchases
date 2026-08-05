package domain

type GroupDetails struct {
	Id          int
	Name        string
	AdminUserID int
	AdminUser   string
}

type UpdateGroup struct {
	Name        *string
	AdminUserID *int
}
