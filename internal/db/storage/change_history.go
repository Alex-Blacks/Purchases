package storage

type ChangeHistoryRepo struct{}

func NewHistoryRepo() *ChangeHistoryRepo {
	return &ChangeHistoryRepo{}
}
