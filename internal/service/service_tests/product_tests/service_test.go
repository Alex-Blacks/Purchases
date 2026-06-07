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
		conflict     bool
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
			conflict:  true,
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
			svc := service.NewServiceProduct(txMock, repoMock)
			if tt.conflict {
				svc.CreateProductAlias(context.Background(), tt.deleteID, "test_alias")
			}

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

func TestProduct_ListProducts(t *testing.T) {
	tests := []struct {
		name string
		seed map[int]ProductInput
		want []domain.ProductDetails
	}{
		{
			name: "success",
			seed: map[int]ProductInput{
				2: {ID: 2, Title: "B", Unit: "kg", CategoryID: 1},
				1: {ID: 1, Title: "A", Unit: "kg", CategoryID: 1},
				3: {ID: 3, Title: "C", Unit: "kg", CategoryID: 1},
			},
			want: []domain.ProductDetails{
				{ID: 1, Title: "A", Unit: "kg", Category: "Category_1"},
				{ID: 2, Title: "B", Unit: "kg", Category: "Category_1"},
				{ID: 3, Title: "C", Unit: "kg", Category: "Category_1"},
			},
		},
		{
			name: "empty",
			seed: map[int]ProductInput{},
			want: []domain.ProductDetails{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockProduct()
			repoMock.products = tt.seed
			svc := service.NewServiceProduct(txMock, repoMock)

			products, err := svc.ListProducts(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(products) != len(tt.want) {
				t.Fatalf("expected %d items, got %d", len(tt.want), len(products))
			}
			for i := range products {
				if products[i] != tt.want[i] {
					t.Fatalf("at index %d expected %+v, got %+v", i, tt.want[i], products[i])
				}
			}
		})
	}
}

func TestProduct_CreateProductAlias(t *testing.T) {
	tests := []struct {
		name           string
		seedProducts   map[int]ProductInput
		seedAliases    map[int]AliasInput
		productID      int
		alias          string
		wantErr        bool
		wantErrIs      error
		checkTxFunc    func(t *testing.T, tx *MockTx)
		checkAliasFunc func(t *testing.T, repo *MockProduct, alias string)
	}{
		{
			name: "success",
			seedProducts: map[int]ProductInput{
				1: {ID: 1, Title: "Яблоки", Unit: "кг", CategoryID: 1},
			},
			productID: 1,
			alias:     "apple",
			wantErr:   false,
			checkTxFunc: func(t *testing.T, tx *MockTx) {
				if !tx.committed {
					t.Error("transaction not committed")
				}
				if tx.rolledBack {
					t.Error("transaction should not be rolled back")
				}
			},
			checkAliasFunc: func(t *testing.T, repo *MockProduct, alias string) {
				found := false
				for _, a := range repo.aliases {
					if a.Alias == alias {
						found = true
						break
					}
				}
				if !found {
					t.Error("alias not found in repo")
				}
			},
		},
		{
			name:         "product not found",
			seedProducts: map[int]ProductInput{},
			productID:    1,
			alias:        "apple",
			wantErr:      true,
			wantErrIs:    domain.ErrConflict,
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
			name: "alias already exists",
			seedProducts: map[int]ProductInput{
				1: {ID: 1, Title: "Яблоки", Unit: "кг", CategoryID: 1},
			},
			seedAliases: map[int]AliasInput{
				1: {ID: 1, ProductID: 1, Alias: "apple"},
			},
			productID: 1,
			alias:     "apple",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockProduct()
			for id, prod := range tt.seedProducts {
				repoMock.products[id] = prod
				if id >= repoMock.nextProductID {
					repoMock.nextProductID = id + 1
				}
			}
			for id, alias := range tt.seedAliases {
				repoMock.aliases[id] = alias
				if id >= repoMock.nextAliasID {
					repoMock.nextAliasID = id + 1
				}
			}
			svc := service.NewServiceProduct(txMock, repoMock)

			aliasDetails, err := svc.CreateProductAlias(context.Background(), tt.productID, tt.alias)

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
				if aliasDetails.Alias != tt.alias {
					t.Fatalf("expected alias %q, got %q", tt.alias, aliasDetails.Alias)
				}
				if aliasDetails.ID == 0 {
					t.Error("alias ID not assigned")
				}
				if aliasDetails.Product != tt.seedProducts[tt.productID].Title {
					t.Errorf("expected product title %q, got %q", tt.seedProducts[tt.productID].Title, aliasDetails.Product)
				}
			}

			if tt.checkTxFunc != nil {
				tt.checkTxFunc(t, txMock)
			}
			if tt.checkAliasFunc != nil {
				tt.checkAliasFunc(t, repoMock, tt.alias)
			}
		})
	}
}

func TestProduct_GetProductAlias(t *testing.T) {
	tests := []struct {
		name         string
		seedAliases  map[int]AliasInput
		seedProducts map[int]ProductInput
		id           int
		wantErr      bool
		wantErrIs    error
		want         domain.ProductAliasDetails
	}{
		{
			name: "success",
			seedProducts: map[int]ProductInput{
				1: {ID: 1, Title: "Яблоки", Unit: "кг", CategoryID: 1},
			},
			seedAliases: map[int]AliasInput{
				10: {ID: 10, ProductID: 1, Alias: "apple"},
			},
			id:      10,
			wantErr: false,
			want: domain.ProductAliasDetails{
				ID:      10,
				Product: "Яблоки",
				Alias:   "apple",
			},
		},
		{
			name:        "not found",
			seedAliases: map[int]AliasInput{},
			id:          1,
			wantErr:     true,
			wantErrIs:   domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockProduct()
			for id, prod := range tt.seedProducts {
				repoMock.products[id] = prod
			}
			for id, alias := range tt.seedAliases {
				repoMock.aliases[id] = alias
			}
			svc := service.NewServiceProduct(txMock, repoMock)

			got, err := svc.GetProductAlias(context.Background(), tt.id)

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

func TestProduct_DeleteProductAlias(t *testing.T) {
	tests := []struct {
		name         string
		seedAliases  map[int]AliasInput
		deleteID     int
		wantErr      bool
		wantErrIs    error
		checkRemains func(t *testing.T, repo *MockProduct)
		checkTxFunc  func(t *testing.T, tx *MockTx)
	}{
		{
			name: "success",
			seedAliases: map[int]AliasInput{
				1: {ID: 1, ProductID: 1, Alias: "apple"},
			},
			deleteID: 1,
			wantErr:  false,
			checkRemains: func(t *testing.T, repo *MockProduct) {
				if _, ok := repo.aliases[1]; ok {
					t.Error("alias still exists after deletion")
				}
			},
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
			name:         "not found",
			seedAliases:  map[int]AliasInput{},
			deleteID:     1,
			wantErr:      true,
			wantErrIs:    domain.ErrNotFound,
			checkRemains: func(t *testing.T, repo *MockProduct) {},
			checkTxFunc: func(t *testing.T, tx *MockTx) {
				if tx.committed {
					t.Error("transaction committed despite error")
				}
				if !tx.rolledBack {
					t.Error("transaction should be rolled back")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockProduct()
			for id, alias := range tt.seedAliases {
				repoMock.aliases[id] = alias
			}
			svc := service.NewServiceProduct(txMock, repoMock)

			err := svc.DeleteProductAlias(context.Background(), tt.deleteID)

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
			if tt.checkTxFunc != nil {
				tt.checkTxFunc(t, txMock)
			}
		})
	}
}

func TestProduct_ListProductAliases(t *testing.T) {
	tests := []struct {
		name         string
		seedProducts map[int]ProductInput
		seedAliases  map[int]AliasInput
		productID    int
		want         []domain.ProductAliasDetails
	}{
		{
			name: "success",
			seedProducts: map[int]ProductInput{
				1: {ID: 1, Title: "Яблоки", Unit: "кг", CategoryID: 1},
			},
			seedAliases: map[int]AliasInput{
				3: {ID: 3, ProductID: 1, Alias: "apple"},
				1: {ID: 1, ProductID: 1, Alias: "apfel"},
				2: {ID: 2, ProductID: 1, Alias: "pomme"},
			},
			productID: 1,
			want: []domain.ProductAliasDetails{
				{ID: 1, Product: "Яблоки", Alias: "apfel"},
				{ID: 2, Product: "Яблоки", Alias: "pomme"},
				{ID: 3, Product: "Яблоки", Alias: "apple"},
			},
		},
		{
			name:         "no aliases",
			seedProducts: map[int]ProductInput{1: {ID: 1, Title: "Яблоки", Unit: "кг", CategoryID: 1}},
			seedAliases:  map[int]AliasInput{},
			productID:    1,
			want:         []domain.ProductAliasDetails{},
		},
		{
			name:         "product not found - empty list",
			seedProducts: map[int]ProductInput{},
			productID:    999,
			want:         []domain.ProductAliasDetails{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockProduct()
			for id, prod := range tt.seedProducts {
				repoMock.products[id] = prod
			}
			for id, alias := range tt.seedAliases {
				repoMock.aliases[id] = alias
				if id >= repoMock.nextAliasID {
					repoMock.nextAliasID = id + 1
				}
			}
			svc := service.NewServiceProduct(txMock, repoMock)

			got, err := svc.ListProductAliases(context.Background(), tt.productID)
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

func TestProduct_DeleteAllProductAliases(t *testing.T) {
	tests := []struct {
		name         string
		seedProducts map[int]ProductInput
		seedAliases  map[int]AliasInput
		productID    int
		wantErr      bool
		wantErrIs    error
		checkRemains func(t *testing.T, repo *MockProduct)
		checkTxFunc  func(t *testing.T, tx *MockTx)
	}{
		{
			name: "success - delete multiple aliases",
			seedProducts: map[int]ProductInput{
				1: {ID: 1, Title: "Яблоки", Unit: "кг", CategoryID: 1},
			},
			seedAliases: map[int]AliasInput{
				1: {ID: 1, ProductID: 1, Alias: "apple"},
				2: {ID: 2, ProductID: 1, Alias: "apfel"},
			},
			productID: 1,
			wantErr:   false,
			checkRemains: func(t *testing.T, repo *MockProduct) {
				for _, a := range repo.aliases {
					if a.ProductID == 1 {
						t.Error("alias still exists for product after DeleteAll")
					}
				}
			},
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
			name: "no aliases to delete - returns ErrNotFound",
			seedProducts: map[int]ProductInput{
				1: {ID: 1, Title: "Яблоки", Unit: "кг", CategoryID: 1},
			},
			seedAliases: map[int]AliasInput{},
			productID:   1,
			wantErr:     true,
			wantErrIs:   domain.ErrNotFound,
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
			name:         "product not found",
			seedProducts: map[int]ProductInput{},
			productID:    1,
			wantErr:      true,
			wantErrIs:    domain.ErrNotFound,
			checkTxFunc: func(t *testing.T, tx *MockTx) {
				if tx.committed {
					t.Error("transaction committed despite error")
				}
				if !tx.rolledBack {
					t.Error("transaction should be rolled back")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockProduct()
			for id, prod := range tt.seedProducts {
				repoMock.products[id] = prod
			}
			for id, alias := range tt.seedAliases {
				repoMock.aliases[id] = alias
			}
			svc := service.NewServiceProduct(txMock, repoMock)

			err := svc.DeleteAllProductAliases(context.Background(), tt.productID)

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
			if tt.checkTxFunc != nil {
				tt.checkTxFunc(t, txMock)
			}
		})
	}
}

func TestProduct_FindProductByAlias(t *testing.T) {
	tests := []struct {
		name         string
		seedProducts map[int]ProductInput
		seedAliases  map[int]AliasInput
		alias        string
		wantErr      bool
		wantErrIs    error
		wantTitle    string
	}{
		{
			name: "success",
			seedProducts: map[int]ProductInput{
				1: {ID: 1, Title: "Яблоки", Unit: "кг", CategoryID: 1},
			},
			seedAliases: map[int]AliasInput{
				1: {ID: 1, ProductID: 1, Alias: "apple"},
			},
			alias:     "apple",
			wantErr:   false,
			wantTitle: "Яблоки",
		},
		{
			name: "alias not found",
			seedProducts: map[int]ProductInput{
				1: {ID: 1, Title: "Яблоки", Unit: "кг", CategoryID: 1},
			},
			seedAliases: map[int]AliasInput{},
			alias:       "unknown",
			wantErr:     true,
			wantErrIs:   domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockProduct()
			for id, prod := range tt.seedProducts {
				repoMock.products[id] = prod
			}
			for id, alias := range tt.seedAliases {
				repoMock.aliases[id] = alias
			}
			svc := service.NewServiceProduct(txMock, repoMock)

			title, err := svc.FindProductByAlias(context.Background(), tt.alias)

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
				if title != tt.wantTitle {
					t.Fatalf("expected title %q, got %q", tt.wantTitle, title)
				}
			}
		})
	}
}
