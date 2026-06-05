package product_tests

import (
	"context"
	"errors"
	"testing"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/service"
)

func TestProduct_CreateProduct(t *testing.T) {
	tests := []struct {
		name  string
		seed  map[int]ProductInput
		input struct {
			title, unit string
			categoryID  int
		}
		wantErr     bool
		wantErrIs   error
		checkTxFunc func(t *testing.T, tx *MockTx)
	}{
		{
			name: "success",
			seed: map[int]ProductInput{1: {ID: 1, Title: "Груши", Unit: "кг", CategoryID: 1}},
			input: struct {
				title      string
				unit       string
				categoryID int
			}{"Яблоки", "кг", 1},
			wantErr: false,
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
			name: "already exists",
			seed: map[int]ProductInput{1: {ID: 1, Title: "Яблоки", Unit: "кг", CategoryID: 1}},
			input: struct {
				title      string
				unit       string
				categoryID int
			}{"Яблоки", "кг", 1},
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
			name: "conflict",
			seed: map[int]ProductInput{},
			input: struct {
				title      string
				unit       string
				categoryID int
			}{"Груши", "кг", 99},
			wantErr:   true,
			wantErrIs: domain.ErrConflict,
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
			name: "case sensitive - different case is allowed",
			seed: map[int]ProductInput{1: {ID: 1, Title: "Яблоки", Unit: "кг", CategoryID: 1}},
			input: struct {
				title      string
				unit       string
				categoryID int
			}{"яблоки", "кг", 1},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockProduct()
			repoMock.products = tt.seed
			maxID := 0
			for id := range tt.seed {
				if id > maxID {
					maxID = id
				}
			}
			repoMock.nextProductID = maxID + 1
			svc := service.NewServiceProduct(txMock, repoMock)

			prod, err := svc.CreateProduct(context.Background(), tt.input.title, tt.input.unit, tt.input.categoryID)

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
				if prod.Title != tt.input.title {
					t.Fatalf("expected name %q, got %q", tt.input.title, prod.Title)
				}
				// проверяем, что ID проставлен
				if prod.ID == 0 {
					t.Error("ID not assigned")
				}
			}

			if tt.checkTxFunc != nil {
				tt.checkTxFunc(t, txMock)
			}
		})
	}
}

func TestProduct_GetProduct(t *testing.T) {
	tests := []struct {
		name      string
		seed      map[int]ProductInput
		id        int
		wantErr   bool
		wantErrIs error
		want      domain.ProductDetails
	}{
		{
			name:    "success",
			seed:    map[int]ProductInput{1: {ID: 1, Title: "Груши", Unit: "кг", CategoryID: 1}},
			id:      1,
			wantErr: false,
			want:    domain.ProductDetails{ID: 1, Title: "Груши", Unit: "кг", Category: "Category_1"},
		},
		{
			name:      "not found",
			seed:      map[int]ProductInput{},
			id:        1,
			wantErr:   true,
			wantErrIs: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockProduct()
			repoMock.products = tt.seed
			svc := service.NewServiceProduct(txMock, repoMock)

			prod, err := svc.GetProduct(context.Background(), tt.id)

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
				if prod != tt.want {
					t.Fatalf("expected %+v, got %+v", tt.want, prod)
				}
			}
		})
	}
}

func TestProduct_DeleteProduct(t *testing.T) {
	tests := []struct {
		name         string
		seed         map[int]ProductInput
		deleteID     int
		wantErr      bool
		wantErrIs    error
		checkRemains func(t *testing.T, repo *MockProduct)
	}{
		{
			name:     "success",
			seed:     map[int]ProductInput{1: {ID: 1, Title: "Груши", Unit: "кг", CategoryID: 1}},
			deleteID: 1,
			wantErr:  false,
			checkRemains: func(t *testing.T, repo *MockProduct) {
				if _, ok := repo.products[1]; ok {
					t.Error("product still exists after deletion")
				}
			},
		},
		{
			name:         "not found",
			seed:         map[int]ProductInput{},
			deleteID:     1,
			wantErr:      true,
			wantErrIs:    domain.ErrNotFound,
			checkRemains: func(t *testing.T, repo *MockProduct) {},
		},
		{
			name:      "conflict - foreign key exists",
			seed:      map[int]ProductInput{1: {ID: 1, Title: "Груши", Unit: "кг", CategoryID: 1}},
			deleteID:  1,
			wantErr:   true,
			wantErrIs: domain.ErrConflict,
			checkRemains: func(t *testing.T, repo *MockProduct) {
				if _, ok := repo.products[1]; !ok {
					t.Error("product was deleted despite conflict")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockProduct()
			repoMock.products = tt.seed
			maxID := 0
			for id := range tt.seed {
				if id > maxID {
					maxID = id
				}
			}
			repoMock.nextProductID = maxID + 1
			svc := service.NewServiceProduct(txMock, repoMock)

			svc.CreateProductAlias(context.Background(), tt.deleteID, "test_alias")

			err := svc.DeleteProduct(context.Background(), tt.deleteID)

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
