package user_tests

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
	m.committed = false
	m.rolledBack = false
	return m, nil
}

type MockUser struct {
	users  map[int]domain.User
	nextID int
}

func NewMockUser() *MockUser {
	return &MockUser{
		users:  make(map[int]domain.User),
		nextID: 1,
	}
}

func (m *MockUser) AddUser(user domain.User) {
	user.ID = m.nextID
	m.nextID++
	m.users[user.ID] = user
}

func (m *MockUser) CreateUser(ctx context.Context, q domain.Querier, name, passwordHash, email, role, status string) (domain.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return domain.User{}, domain.ErrAlreadyExists
		}
	}
	id := m.nextID
	m.nextID++
	user := domain.User{
		ID:           id,
		Name:         name,
		PasswordHash: passwordHash,
		Email:        email,
		Role:         role,
		Status:       status,
	}
	m.users[id] = user
	return user, nil
}

func (m *MockUser) GetUserByID(ctx context.Context, q domain.Querier, userID int) (domain.User, error) {
	user, ok := m.users[userID]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

func (m *MockUser) GetUserByEmail(ctx context.Context, q domain.Querier, email string) (domain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return domain.User{}, domain.ErrNotFound
}

func (m *MockUser) UpdateUser(ctx context.Context, q domain.Querier, userID int, updateUser domain.UpdateUser) (domain.User, error) {
	user, ok := m.users[userID]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	if updateUser.Name != nil {
		user.Name = *updateUser.Name
	}
	if updateUser.Password != nil {
		user.PasswordHash = *updateUser.Password
	}
	if updateUser.Email != nil {
		user.Email = *updateUser.Email
	}
	if updateUser.Role != nil {
		user.Role = *updateUser.Role
	}
	if updateUser.Status != nil {
		user.Status = *updateUser.Status
	}
	m.users[userID] = user
	return user, nil
}

func (m *MockUser) DeleteUser(ctx context.Context, q domain.Querier, userID int) error {
	if _, ok := m.users[userID]; !ok {
		return domain.ErrNotFound
	}
	delete(m.users, userID)
	return nil
}

func (m *MockUser) ListUsers(ctx context.Context, q domain.Querier) ([]domain.User, error) {
	result := make([]domain.User, 0, len(m.users))
	for _, user := range m.users {
		result = append(result, user)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}
