package service

import (
	"github.com/Alex-Blacks/Purchases/internal/domain"
)

type ServiceStore struct {
	*GenericService[domain.StoreDetails, domain.StoreRepository]
}

func NewServiceStore(st domain.Storage, repo domain.StoreRepository) *ServiceStore {
	return &ServiceStore{
		GenericService: &GenericService[domain.StoreDetails, domain.StoreRepository]{
			BaseService: &BaseService{storage: st},
			repo:        repo,
		},
	}
}
