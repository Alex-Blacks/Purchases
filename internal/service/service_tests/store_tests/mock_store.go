package store_tests

import (
	"context"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type MockTx struct{}

func NewMockTx() *MockTx {
	return &MockTx{}
}

func (m *MockTx) Commit(ctx context.Context) error {
	return nil
}

func (m *MockTx) Rollback(ctx context.Context) error {
	return nil
}

func (m *MockTx) Exec(
	ctx context.Context,
	sql string,
	args ...any,
) (pgconn.CommandTag, error) {
	panic("unexpected call")
}

func (m *MockTx) Query(
	ctx context.Context,
	sql string,
	args ...any,
) (pgx.Rows, error) {
	panic("unexpected call")
}

func (m *MockTx) QueryRow(
	ctx context.Context,
	sql string,
	args ...any,
) pgx.Row {
	panic("unexpected call")
}

func (ms *MockTx) BeginTx(ctx context.Context) (domain.Tx, error) {
	return &MockTx{}, nil
}

type MockStore struct {
	data   map[int]domain.Store
	nextID int
}

func NewMockStore() *MockStore {
	return &MockStore{
		data:   make(map[int]domain.Store),
		nextID: 1,
	}
}

func (ms *MockStore) CreateStore(ctx context.Context, q domain.Querier, name string) (domain.Store, error) {
	for _, store := range ms.data {
		if store.Name == name {
			return domain.Store{}, domain.ErrAlreadyExists
		}
	}

	id := ms.nextID
	ms.nextID++

	store := domain.Store{
		ID:   id,
		Name: name,
	}

	ms.data[id] = store

	return store, nil
}

func (ms *MockStore) GetStore(ctx context.Context, q domain.Querier, id int) (domain.Store, error) {
	store, ok := ms.data[id]
	if !ok {
		return domain.Store{}, domain.ErrNotFound
	}

	return store, nil
}

func (ms *MockStore) DeleteStore(ctx context.Context, q domain.Querier, id int) error {
	_, ok := ms.data[id]
	if !ok {
		return domain.ErrNotFound
	}

	delete(ms.data, id)
	return nil
}

func (ms *MockStore) ListStores(ctx context.Context, q domain.Querier) ([]domain.Store, error) {
	result := make([]domain.Store, 0)

	for _, store := range ms.data {
		result = append(result, store)
	}

	return result, nil
}

func NewTestServiceStore(seed map[int]domain.Store) (*service.ServiceStore, *MockStore) {
	storageTx := NewMockTx()
	storageRepo := NewMockStore()
	storageRepo.data = seed
	storageRepo.nextID = 1
	svc := service.NewServiceStore(storageTx, storageRepo)

	return svc, storageRepo
}
