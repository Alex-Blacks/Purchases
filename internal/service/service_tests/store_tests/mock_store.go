package store_tests

import (
	"context"
	"errors"
	"sort"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type MockTx struct {
	committed  bool
	rolledBack bool
}

func NewMockTx() *MockTx {
	return &MockTx{}
}

func (m *MockTx) Commit(ctx context.Context) error {
	if m.committed || m.rolledBack {
		return errors.New("transaction already finished")
	}
	m.committed = true
	return nil
}

func (m *MockTx) Rollback(ctx context.Context) error {
	if m.committed || m.rolledBack {
		return errors.New("transaction already finished")
	}
	m.rolledBack = true
	return nil
}

func (m *MockTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	panic("unexpected call")
}

func (m *MockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	panic("unexpected call")
}

func (m *MockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	panic("unexpected call")
}

func (m *MockTx) BeginTx(ctx context.Context) (domain.Tx, error) {
	return m, nil
}

type MockStore struct {
	data        map[int]domain.Store
	nextID      int
	foreignKeys map[int]bool
}

func NewMockStore() *MockStore {
	return &MockStore{
		data:        make(map[int]domain.Store),
		nextID:      1,
		foreignKeys: make(map[int]bool),
	}
}

func (mc *MockStore) AddForeignKey(id int) {
	mc.foreignKeys[id] = true
}

func (ms *MockStore) CreateStore(ctx context.Context, q domain.Querier, name string) (domain.Store, error) {
	for _, store := range ms.data {
		if store.Name == name {
			return domain.Store{}, domain.ErrAlreadyExists
		}
	}

	id := ms.nextID
	ms.nextID++
	store := domain.Store{ID: id, Name: name}
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
	if ms.foreignKeys[id] {
		return domain.ErrConflict
	}

	delete(ms.data, id)
	return nil
}

func (ms *MockStore) ListStores(ctx context.Context, q domain.Querier) ([]domain.Store, error) {
	result := make([]domain.Store, 0, len(ms.data))

	for _, store := range ms.data {
		result = append(result, store)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}
