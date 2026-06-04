package category_tests

import (
	"context"
	"errors"
	"testing"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/service"
)

func TestCategory_CreateCategory(t *testing.T) {
	tests := []struct {
		name        string
		seed        map[int]domain.Category
		inputName   string
		wantErr     bool
		wantErrIs   error
		checkTxFunc func(t *testing.T, tx *MockTx)
	}{
		{
			name:      "success",
			seed:      map[int]domain.Category{},
			inputName: "Test1",
			wantErr:   false,
			checkTxFunc: func(t *testing.T, tx *MockTx) {
				if !tx.committed {
					t.Error("transaction not committed")
				}
				if tx.rolledBack {
					t.Error("transaction should not be rolled back")
				}
			},
		},
		{
			name:      "already exists",
			seed:      map[int]domain.Category{1: {ID: 1, Name: "Test1"}},
			inputName: "Test1",
			wantErr:   true,
			wantErrIs: domain.ErrAlreadyExists,
			checkTxFunc: func(t *testing.T, tx *MockTx) {
				if tx.committed {
					t.Error("transaction committed despite error")
				}
				if !tx.rolledBack {
					t.Error("transaction should be rolled back")
				}
			},
		},
		{
			name:      "case sensitive - different case is allowed",
			seed:      map[int]domain.Category{1: {ID: 1, Name: "Test1"}},
			inputName: "test1",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockCategory()
			repoMock.data = tt.seed
			repoMock.nextID = 1
			svc := service.NewServiceCategory(txMock, repoMock)

			cat, err := svc.CreateCategory(context.Background(), tt.inputName)

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
				if cat.Name != tt.inputName {
					t.Fatalf("expected name %q, got %q", tt.inputName, cat.Name)
				}
				// проверяем, что ID проставлен
				if cat.ID == 0 {
					t.Error("ID not assigned")
				}
			}

			if tt.checkTxFunc != nil {
				tt.checkTxFunc(t, txMock)
			}
		})
	}
}

func TestCategory_GetCategory(t *testing.T) {
	tests := []struct {
		name      string
		seed      map[int]domain.Category
		id        int
		wantErr   bool
		wantErrIs error
		want      domain.Category
	}{
		{
			name: "success",
			seed: map[int]domain.Category{1: {ID: 1, Name: "Test1"}},
			id:   1,
			want: domain.Category{ID: 1, Name: "Test1"},
		},
		{
			name:      "not found",
			seed:      map[int]domain.Category{},
			id:        1,
			wantErr:   true,
			wantErrIs: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockCategory()
			repoMock.data = tt.seed
			svc := service.NewServiceCategory(txMock, repoMock)

			got, err := svc.GetCategory(context.Background(), tt.id)

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
				if got != tt.want {
					t.Fatalf("expected %+v, got %+v", tt.want, got)
				}
			}
		})
	}
}

func TestCategory_DeleteCategory(t *testing.T) {
	tests := []struct {
		name         string
		seed         map[int]domain.Category
		foreignKeys  []int // id категорий, на которые есть внешние ссылки
		deleteID     int
		wantErr      bool
		wantErrIs    error
		checkRemains func(t *testing.T, repo *MockCategory) // проверка, что запись осталась/удалена
	}{
		{
			name:     "success",
			seed:     map[int]domain.Category{1: {ID: 1, Name: "Test1"}},
			deleteID: 1,
			wantErr:  false,
			checkRemains: func(t *testing.T, repo *MockCategory) {
				if _, ok := repo.data[1]; ok {
					t.Error("category still exists after deletion")
				}
			},
		},
		{
			name:      "not found",
			seed:      map[int]domain.Category{},
			deleteID:  1,
			wantErr:   true,
			wantErrIs: domain.ErrNotFound,
			checkRemains: func(t *testing.T, repo *MockCategory) {
				// ничего не должно измениться
			},
		},
		{
			name:        "conflict - foreign key exists",
			seed:        map[int]domain.Category{1: {ID: 1, Name: "Test1"}},
			foreignKeys: []int{1},
			deleteID:    1,
			wantErr:     true,
			wantErrIs:   domain.ErrConflict,
			checkRemains: func(t *testing.T, repo *MockCategory) {
				if _, ok := repo.data[1]; !ok {
					t.Error("category was deleted despite conflict")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockCategory()
			repoMock.data = tt.seed
			for _, id := range tt.foreignKeys {
				repoMock.AddForeignKey(id)
			}
			svc := service.NewServiceCategory(txMock, repoMock)

			err := svc.DeleteCategory(context.Background(), tt.deleteID)

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

			// проверки транзакции
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

func TestCategory_ListCategories(t *testing.T) {
	tests := []struct {
		name string
		seed map[int]domain.Category
		want []domain.Category
	}{
		{
			name: "success",
			seed: map[int]domain.Category{
				2: {ID: 2, Name: "B"},
				1: {ID: 1, Name: "A"},
				3: {ID: 3, Name: "C"},
			},
			want: []domain.Category{
				{ID: 1, Name: "A"},
				{ID: 2, Name: "B"},
				{ID: 3, Name: "C"},
			},
		},
		{
			name: "empty",
			seed: map[int]domain.Category{},
			want: []domain.Category{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockCategory()
			repoMock.data = tt.seed
			svc := service.NewServiceCategory(txMock, repoMock)

			got, err := svc.ListCategories(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("expected %d items, got %d", len(tt.want), len(got))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("at index %d expected %+v, got %+v", i, tt.want[i], got[i])
				}
			}
		})
	}
}
