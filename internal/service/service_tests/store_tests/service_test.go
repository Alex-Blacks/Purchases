package store_tests

import (
	"context"
	"errors"
	"testing"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/service"
)

func TestStore_CreateStore(t *testing.T) {
	tests := []struct {
		name        string
		seed        map[int]domain.Store
		inputName   string
		wantErr     bool
		wantErrIs   error
		checkTxFunc func(t *testing.T, tx *MockTx)
	}{
		{
			name:      "success",
			seed:      map[int]domain.Store{},
			inputName: "Test1",
			wantErr:   false,
			checkTxFunc: func(t *testing.T, tx *MockTx) {
				if !tx.committed {
					t.Error("transaction not commited")
				}
				if tx.rolledBack {
					t.Error("transaction should not be rolled back")
				}
			},
		},
		{
			name:      "already exists",
			seed:      map[int]domain.Store{1: {ID: 1, Name: "Test1"}},
			inputName: "Test1",
			wantErr:   true,
			wantErrIs: domain.ErrAlreadyExists,
			checkTxFunc: func(t *testing.T, tx *MockTx) {
				if tx.committed {
					t.Error("transaction commited despite error")
				}
				if !tx.rolledBack {
					t.Error("transaction should be rolled back")
				}
			},
		},
		{
			name:      "case sensitive - different case is allowed",
			seed:      map[int]domain.Store{1: {ID: 1, Name: "Test1"}},
			inputName: "test1",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockStore()
			repoMock.data = tt.seed
			repoMock.nextID = 1
			svc := service.NewServiceStore(txMock, repoMock)

			store, err := svc.CreateStore(context.Background(), tt.inputName)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
					return
				}
			} else {
				if err != nil {
					t.Fatal("unexpected error")
				}

				if store.Name != tt.inputName {
					t.Fatalf("expected name %q, got %q", tt.inputName, store.Name)
				}
				if store.ID == 0 {
					t.Error("ID not assigned")
				}
			}

			if tt.checkTxFunc != nil {
				tt.checkTxFunc(t, txMock)
			}

		})
	}
}

func TestStore_GetStore(t *testing.T) {
	tests := []struct {
		name      string
		seed      map[int]domain.Store
		id        int
		wantErr   bool
		wantErrIs error
		want      domain.Store
	}{
		{
			name:    "success",
			seed:    map[int]domain.Store{1: {ID: 1, Name: "Test1"}},
			id:      1,
			wantErr: false,
			want:    domain.Store{ID: 1, Name: "Test1"},
		},
		{
			name:      "not found",
			seed:      make(map[int]domain.Store),
			id:        1,
			wantErr:   true,
			wantErrIs: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockStore()
			repoMock.data = tt.seed
			svc := service.NewServiceStore(txMock, repoMock)

			store, err := svc.GetStore(context.Background(), tt.id)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if store != tt.want {
					t.Fatalf("expected %+v, got %+v", tt.want, store)
				}
			}

		})
	}
}

func TestStore_DeleteStore(t *testing.T) {
	tests := []struct {
		name         string
		seed         map[int]domain.Store
		foreignKeys  []int
		deleteID     int
		wantErr      bool
		wantErrIs    error
		checkRemains func(t *testing.T, repo *MockStore)
	}{
		{
			name:     "success",
			seed:     map[int]domain.Store{1: {ID: 1, Name: "Test1"}},
			deleteID: 1,
			wantErr:  false,
			checkRemains: func(t *testing.T, repo *MockStore) {
				if _, ok := repo.data[1]; ok {
					t.Error("category still exists after deletion")
				}
			},
		},
		{
			name:         "not found",
			seed:         make(map[int]domain.Store),
			deleteID:     1,
			wantErr:      true,
			wantErrIs:    domain.ErrNotFound,
			checkRemains: func(t *testing.T, repo *MockStore) {},
		},
		{
			name:        "conflict - foreign key exists",
			seed:        map[int]domain.Store{1: {ID: 1, Name: "Test1"}},
			foreignKeys: []int{1},
			deleteID:    1,
			wantErr:     true,
			wantErrIs:   domain.ErrConflict,
			checkRemains: func(t *testing.T, repo *MockStore) {
				if _, ok := repo.data[1]; !ok {
					t.Error("category was deleted despite conflict")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockStore()
			repoMock.data = tt.seed
			for _, id := range tt.foreignKeys {
				repoMock.AddForeignKey(id)
			}
			svc := service.NewServiceStore(txMock, repoMock)

			err := svc.DeleteStore(context.Background(), tt.deleteID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			if tt.checkRemains != nil {
				tt.checkRemains(t, repoMock)
			}

			if tt.wantErr {
				if !txMock.rolledBack {
					t.Error("transaction should be rolled back on error")
				}
				if txMock.committed {
					t.Error("transaction committed despite error")
				}
			} else {
				if !txMock.committed {
					t.Error("transaction not committed on success")
				}
				if txMock.rolledBack {
					t.Error("transaction rolled back on success")
				}
			}
		})
	}
}

func TestStore_ListStores(t *testing.T) {
	tests := []struct {
		name string
		seed map[int]domain.Store
		want []domain.Store
	}{
		{
			name: "success",
			seed: map[int]domain.Store{
				2: {ID: 2, Name: "B"},
				1: {ID: 1, Name: "A"},
				3: {ID: 3, Name: "C"},
			},
			want: []domain.Store{
				{ID: 1, Name: "A"},
				{ID: 2, Name: "B"},
				{ID: 3, Name: "C"},
			},
		},
		{
			name: "empty",
			seed: map[int]domain.Store{},
			want: []domain.Store{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockStore()
			repoMock.data = tt.seed
			svc := service.NewServiceStore(txMock, repoMock)

			stores, err := svc.ListStores(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(stores) != len(tt.want) {
				t.Fatalf("expected %d items, got %d", len(tt.want), len(stores))
			}
			for i := range stores {
				if stores[i] != tt.want[i] {
					t.Fatalf("at index %d expected %+v, got %+v", i, tt.want[i], stores[i])
				}
			}
		})
	}
}
