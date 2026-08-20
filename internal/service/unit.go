package service

import (
	"github.com/Alex-Blacks/Purchases/internal/domain"
)

type ServiceUnit struct {
	*GenericService[domain.UnitDetails, domain.UnitRepository]
}

func NewServiceUnit(st domain.Storage, repo domain.UnitRepository) *ServiceUnit {
	return &ServiceUnit{
		GenericService: &GenericService[domain.UnitDetails, domain.UnitRepository]{
			BaseService: &BaseService{storage: st},
			repo:        repo,
		},
	}
}
