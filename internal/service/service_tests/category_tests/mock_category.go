package category_tests

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

func (ms *MockTx) BeginTx(ctx context.Context) (domain.Tx, error) {
	return ms, nil
}

type MockCategory struct {
	data        map[int]domain.Category
	nextID      int
	foreignKeys map[int]bool
}

func NewMockCategory() *MockCategory {
	return &MockCategory{
		data:        make(map[int]domain.Category),
		nextID:      1,
		foreignKeys: make(map[int]bool),
	}
}

func (mc *MockCategory) AddForeignKey(id int) {
	mc.foreignKeys[id] = true
}

func (mc *MockCategory) CreateCategory(ctx context.Context, q domain.Querier, name string) (domain.Category, error) {
	for _, cat := range mc.data {
		if cat.Name == name {
			return domain.Category{}, domain.ErrAlreadyExists
		}
	}
	id := mc.nextID
	mc.nextID++
	cat := domain.Category{ID: id, Name: name}
	mc.data[id] = cat
	return cat, nil
}

func (mc *MockCategory) GetCategory(ctx context.Context, q domain.Querier, id int) (domain.Category, error) {
	cat, ok := mc.data[id]
	if !ok {
		return domain.Category{}, domain.ErrNotFound
	}
	return cat, nil
}

func (mc *MockCategory) DeleteCategory(ctx context.Context, q domain.Querier, id int) error {
	if _, ok := mc.data[id]; !ok {
		return domain.ErrNotFound
	}
	if mc.foreignKeys[id] {
		return domain.ErrConflict
	}
	delete(mc.data, id)
	return nil
}

func (mc *MockCategory) ListCategories(ctx context.Context, q domain.Querier) ([]domain.Category, error) {
	res := make([]domain.Category, 0, len(mc.data))
	for _, cat := range mc.data {
		res = append(res, cat)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})
	return res, nil
}
