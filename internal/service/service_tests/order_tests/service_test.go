package order_tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/policy"
	"github.com/Alex-Blacks/Purchases/internal/service"
)

func TestOrder_CreateOrder(t *testing.T) {
	tests := []struct {
		name        string
		seedOrders  map[int]orderRow
		storeID     int
		userID      int
		actorUserID int
		wantErr     bool
		wantErrIs   error
		checkTxFunc func(t *testing.T, tx *MockTx)
		checkOrder  func(t *testing.T, order domain.OrderWithItemDetails)
	}{
		{
			name:        "success",
			seedOrders:  map[int]orderRow{},
			storeID:     10,
			userID:      1,
			actorUserID: 1,
			wantErr:     false,
			checkTxFunc: func(t *testing.T, tx *MockTx) {
				if !tx.committed {
					t.Error("transaction not committed")
				}
				if tx.rolledBack {
					t.Error("transaction should not be rolled back")
				}
			},
			checkOrder: func(t *testing.T, order domain.OrderWithItemDetails) {
				if order.Order.UserID != 1 {
					t.Errorf("expected userID 1, got %d", order.Order.UserID)
				}
				if order.Order.Store == "" {
					t.Error("store name is empty")
				}
				if order.Order.ID == 0 {
					t.Error("order ID not assigned")
				}
			},
		},
		{
			name:        "store not found (conflict)",
			seedOrders:  map[int]orderRow{},
			storeID:     999,
			userID:      1,
			actorUserID: 1,
			wantErr:     true,
			wantErrIs:   domain.ErrConflict,
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
			repoMock := NewMockOrder()
			for id, ord := range tt.seedOrders {
				repoMock.orders[id] = ord
				if id >= repoMock.nextOrderID {
					repoMock.nextOrderID = id + 1
				}
			}
			if tt.wantErr == false || (tt.wantErr && tt.wantErrIs != domain.ErrConflict) {
				repoMock.AddStore(tt.storeID)
			}
			svc := service.NewServiceOrderItem(txMock, repoMock, repoMock)
			actor := policy.Actor{UserID: tt.actorUserID}

			order, err := svc.CreateOrder(context.Background(), actor, tt.storeID)

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
				if tt.checkOrder != nil {
					tt.checkOrder(t, order)
				}
			}
			if tt.checkTxFunc != nil {
				tt.checkTxFunc(t, txMock)
			}
		})
	}
}

func TestOrder_GetOrder(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name         string
		seedOrders   map[int]orderRow
		seedItems    map[int]orderItemRow
		seedProducts map[int]string
		userID       int
		orderID      int
		actorUserID  int
		wantErr      bool
		wantErrIs    error
		wantOrder    domain.OrderWithItemDetails
	}{
		{
			name: "success with items",
			seedOrders: map[int]orderRow{
				100: {ID: 100, UserID: 1, StoreID: 10, StoreName: "Test Store", CreatedAt: now, UpdatedAt: now},
			},
			seedItems: map[int]orderItemRow{
				1: {ID: 1, OrderID: 100, ProductID: 200, Quantity: 5},
			},
			seedProducts: map[int]string{200: "Apple"},
			userID:       1,
			orderID:      100,
			actorUserID:  1,
			wantErr:      false,
			wantOrder: domain.OrderWithItemDetails{
				Order: domain.OrderDetails{
					ID:         100,
					UserID:     1,
					Store:      "Test Store",
					ItemsCount: 1,
					CreatedAt:  now,
					UpdatedAt:  now,
				},
				Items: []domain.OrderItemDetails{
					{ID: 1, ProductID: 200, Title: "Apple", Quantity: 5},
				},
			},
		},
		{
			name:        "order not found",
			seedOrders:  map[int]orderRow{},
			userID:      1,
			orderID:     999,
			actorUserID: 1,
			wantErr:     true,
			wantErrIs:   domain.ErrNotFound,
		},
		{
			name: "order belongs to another user",
			seedOrders: map[int]orderRow{
				100: {ID: 100, UserID: 2, StoreID: 10, StoreName: "Test Store", CreatedAt: now, UpdatedAt: now},
			},
			userID:      2,
			orderID:     100,
			actorUserID: 1,
			wantErr:     true,
			wantErrIs:   domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockOrder()
			repoMock.orders = tt.seedOrders
			repoMock.orderItems = tt.seedItems
			repoMock.products = tt.seedProducts
			svc := service.NewServiceOrderItem(txMock, repoMock, repoMock)
			actor := policy.Actor{UserID: tt.actorUserID}

			got, err := svc.GetOrder(context.Background(), actor, tt.orderID)

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
				if got.Order.ID != tt.wantOrder.Order.ID {
					t.Errorf("order ID: expected %d, got %d", tt.wantOrder.Order.ID, got.Order.ID)
				}
				if got.Order.UserID != tt.wantOrder.Order.UserID {
					t.Errorf("userID: expected %d, got %d", tt.wantOrder.Order.UserID, got.Order.UserID)
				}
				if got.Order.Store != tt.wantOrder.Order.Store {
					t.Errorf("store: expected %s, got %s", tt.wantOrder.Order.Store, got.Order.Store)
				}
				if len(got.Items) != len(tt.wantOrder.Items) {
					t.Fatalf("items count: expected %d, got %d", len(tt.wantOrder.Items), len(got.Items))
				}
				for i := range got.Items {
					if got.Items[i] != tt.wantOrder.Items[i] {
						t.Errorf("item %d: expected %+v, got %+v", i, tt.wantOrder.Items[i], got.Items[i])
					}
				}
			}
		})
	}
}

func TestOrder_DeleteOrder(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name         string
		seedOrders   map[int]orderRow
		seedItems    map[int]orderItemRow
		userID       int
		orderID      int
		actorUserID  int
		wantErr      bool
		wantErrIs    error
		checkRemains func(t *testing.T, repo *MockOrder)
		checkTxFunc  func(t *testing.T, tx *MockTx)
	}{
		{
			name: "success",
			seedOrders: map[int]orderRow{
				100: {ID: 100, UserID: 1, StoreID: 10, StoreName: "Store", CreatedAt: now, UpdatedAt: now},
			},
			seedItems: map[int]orderItemRow{
				1: {ID: 1, OrderID: 100, ProductID: 200, Quantity: 3},
			},
			userID:      1,
			orderID:     100,
			actorUserID: 1,
			wantErr:     false,
			checkRemains: func(t *testing.T, repo *MockOrder) {
				if _, ok := repo.orders[100]; ok {
					t.Error("order still exists after deletion")
				}
				for _, it := range repo.orderItems {
					if it.OrderID == 100 {
						t.Error("order item still exists after order deletion")
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
			name:        "order not found",
			seedOrders:  map[int]orderRow{},
			userID:      1,
			orderID:     999,
			actorUserID: 1,
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
			name: "order belongs to another user",
			seedOrders: map[int]orderRow{
				100: {ID: 100, UserID: 2, StoreID: 10, StoreName: "Store", CreatedAt: now, UpdatedAt: now},
			},
			userID:      2,
			orderID:     100,
			actorUserID: 1,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockOrder()
			repoMock.orders = tt.seedOrders
			repoMock.orderItems = tt.seedItems
			svc := service.NewServiceOrderItem(txMock, repoMock, repoMock)
			actor := policy.Actor{UserID: tt.actorUserID}

			err := svc.DeleteOrder(context.Background(), actor, tt.orderID)

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

func TestOrder_ListOrders(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name        string
		seedOrders  map[int]orderRow
		seedItems   map[int]orderItemRow
		userID      int
		actorUserID int
		wantErr     bool
		wantErrIs   error
		want        []domain.OrderDetails
	}{
		{
			name: "success multiple orders",
			seedOrders: map[int]orderRow{
				1: {ID: 1, UserID: 1, StoreID: 10, StoreName: "Store A", CreatedAt: now, UpdatedAt: now},
				2: {ID: 2, UserID: 1, StoreID: 20, StoreName: "Store B", CreatedAt: now, UpdatedAt: now},
				3: {ID: 3, UserID: 2, StoreID: 10, StoreName: "Store A", CreatedAt: now, UpdatedAt: now},
			},
			seedItems: map[int]orderItemRow{
				10: {ID: 10, OrderID: 1, ProductID: 100, Quantity: 2},
				11: {ID: 11, OrderID: 1, ProductID: 101, Quantity: 1},
				12: {ID: 12, OrderID: 2, ProductID: 200, Quantity: 3},
			},
			userID:      1,
			actorUserID: 1,
			wantErr:     false,
			want: []domain.OrderDetails{
				{ID: 1, UserID: 1, Store: "Store A", ItemsCount: 2, CreatedAt: now, UpdatedAt: now},
				{ID: 2, UserID: 1, Store: "Store B", ItemsCount: 1, CreatedAt: now, UpdatedAt: now},
			},
		},
		{
			name:        "no orders",
			seedOrders:  map[int]orderRow{},
			userID:      1,
			actorUserID: 1,
			wantErr:     false,
			want:        []domain.OrderDetails{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockOrder()
			repoMock.orders = tt.seedOrders
			repoMock.orderItems = tt.seedItems
			svc := service.NewServiceOrderItem(txMock, repoMock, repoMock)
			actor := policy.Actor{UserID: tt.actorUserID}

			got, err := svc.ListOrders(context.Background(), actor)

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
				if len(got) != len(tt.want) {
					t.Fatalf("expected %d orders, got %d", len(tt.want), len(got))
				}
				for i := range got {
					if got[i].ID != tt.want[i].ID || got[i].UserID != tt.want[i].UserID || got[i].Store != tt.want[i].Store || got[i].ItemsCount != tt.want[i].ItemsCount {
						t.Errorf("order %d: expected %+v, got %+v", i, tt.want[i], got[i])
					}
				}
			}
		})
	}
}

func TestOrder_AddItem(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name         string
		seedOrders   map[int]orderRow
		seedItems    map[int]orderItemRow
		seedProducts map[int]string
		userID       int
		orderID      int
		productID    int
		quantity     int
		actorUserID  int
		wantErr      bool
		wantErrIs    error
		checkItem    func(t *testing.T, item domain.OrderItemDetails)
		checkRemains func(t *testing.T, repo *MockOrder)
		checkTxFunc  func(t *testing.T, tx *MockTx)
	}{
		{
			name: "add new item success",
			seedOrders: map[int]orderRow{
				100: {ID: 100, UserID: 1, StoreID: 10, StoreName: "Store", CreatedAt: now, UpdatedAt: now},
			},
			seedProducts: map[int]string{200: "Orange"},
			userID:       1,
			orderID:      100,
			productID:    200,
			quantity:     3,
			actorUserID:  1,
			wantErr:      false,
			checkItem: func(t *testing.T, item domain.OrderItemDetails) {
				if item.ProductID != 200 {
					t.Errorf("expected productID 200, got %d", item.ProductID)
				}
				if item.Quantity != 3 {
					t.Errorf("expected quantity 3, got %d", item.Quantity)
				}
				if item.Title != "Orange" {
					t.Errorf("expected title Orange, got %s", item.Title)
				}
				if item.ID == 0 {
					t.Error("item ID not assigned")
				}
			},
			checkRemains: func(t *testing.T, repo *MockOrder) {
				found := false
				for _, it := range repo.orderItems {
					if it.OrderID == 100 && it.ProductID == 200 {
						found = true
						if it.Quantity != 3 {
							t.Errorf("stored quantity = %d, expected 3", it.Quantity)
						}
						break
					}
				}
				if !found {
					t.Error("item not stored in mock")
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
			name: "add item when product already exists in order - should update quantity (additive)",
			seedOrders: map[int]orderRow{
				100: {ID: 100, UserID: 1, StoreID: 10, StoreName: "Store", CreatedAt: now, UpdatedAt: now},
			},
			seedItems: map[int]orderItemRow{
				1: {ID: 1, OrderID: 100, ProductID: 200, Quantity: 2},
			},
			seedProducts: map[int]string{200: "Orange"},
			userID:       1,
			orderID:      100,
			productID:    200,
			quantity:     3,
			actorUserID:  1,
			wantErr:      false,
			checkItem: func(t *testing.T, item domain.OrderItemDetails) {
				if item.Quantity != 5 {
					t.Errorf("expected quantity 5, got %d", item.Quantity)
				}
			},
			checkRemains: func(t *testing.T, repo *MockOrder) {
				for _, it := range repo.orderItems {
					if it.OrderID == 100 && it.ProductID == 200 {
						if it.Quantity != 5 {
							t.Errorf("stored quantity = %d, expected 5", it.Quantity)
						}
						return
					}
				}
				t.Error("item not found")
			},
			checkTxFunc: func(t *testing.T, tx *MockTx) {
				if !tx.committed {
					t.Error("transaction not committed")
				}
			},
		},
		{
			name:        "order not found",
			seedOrders:  map[int]orderRow{},
			userID:      1,
			orderID:     100,
			productID:   200,
			quantity:    1,
			actorUserID: 1,
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
			name: "product not found (conflict)",
			seedOrders: map[int]orderRow{
				100: {ID: 100, UserID: 1, StoreID: 10, StoreName: "Store", CreatedAt: now, UpdatedAt: now},
			},
			seedProducts: map[int]string{},
			userID:       1,
			orderID:      100,
			productID:    200,
			quantity:     1,
			actorUserID:  1,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockOrder()
			repoMock.orders = tt.seedOrders
			repoMock.orderItems = tt.seedItems
			repoMock.products = tt.seedProducts
			svc := service.NewServiceOrderItem(txMock, repoMock, repoMock)
			actor := policy.Actor{UserID: tt.actorUserID}

			item, err := svc.AddItem(context.Background(), actor, tt.orderID, tt.productID, tt.quantity)

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
				if tt.checkItem != nil {
					tt.checkItem(t, item)
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

func TestOrder_UpdateItem(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name         string
		seedOrders   map[int]orderRow
		seedItems    map[int]orderItemRow
		seedProducts map[int]string
		userID       int
		orderID      int
		productID    int
		newQuantity  int
		actorUserID  int
		wantErr      bool
		wantErrIs    error
		checkItem    func(t *testing.T, item domain.OrderItemDetails)
		checkTxFunc  func(t *testing.T, tx *MockTx)
	}{
		{
			name: "success",
			seedOrders: map[int]orderRow{
				100: {ID: 100, UserID: 1, StoreID: 10, StoreName: "Store", CreatedAt: now, UpdatedAt: now},
			},
			seedItems: map[int]orderItemRow{
				1: {ID: 1, OrderID: 100, ProductID: 200, Quantity: 2},
			},
			seedProducts: map[int]string{200: "Apple"},
			userID:       1,
			orderID:      100,
			productID:    200,
			newQuantity:  7,
			actorUserID:  1,
			wantErr:      false,
			checkItem: func(t *testing.T, item domain.OrderItemDetails) {
				if item.Quantity != 7 {
					t.Errorf("expected quantity 7, got %d", item.Quantity)
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
			name: "item not found",
			seedOrders: map[int]orderRow{
				100: {ID: 100, UserID: 1, StoreID: 10, StoreName: "Store", CreatedAt: now, UpdatedAt: now},
			},
			seedItems:   map[int]orderItemRow{},
			userID:      1,
			orderID:     100,
			productID:   200,
			newQuantity: 5,
			actorUserID: 1,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockOrder()
			repoMock.orders = tt.seedOrders
			repoMock.orderItems = tt.seedItems
			repoMock.products = tt.seedProducts
			svc := service.NewServiceOrderItem(txMock, repoMock, repoMock)
			actor := policy.Actor{UserID: tt.actorUserID}

			item, err := svc.UpdateItem(context.Background(), actor, tt.orderID, tt.productID, tt.newQuantity)

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
				if tt.checkItem != nil {
					tt.checkItem(t, item)
				}
			}
			if tt.checkTxFunc != nil {
				tt.checkTxFunc(t, txMock)
			}
		})
	}
}

func TestOrder_DeleteItem(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name         string
		seedOrders   map[int]orderRow
		seedItems    map[int]orderItemRow
		userID       int
		orderID      int
		productID    int
		actorUserID  int
		wantErr      bool
		wantErrIs    error
		checkRemains func(t *testing.T, repo *MockOrder)
		checkTxFunc  func(t *testing.T, tx *MockTx)
	}{
		{
			name: "success",
			seedOrders: map[int]orderRow{
				100: {ID: 100, UserID: 1, StoreID: 10, StoreName: "Store", CreatedAt: now, UpdatedAt: now},
			},
			seedItems: map[int]orderItemRow{
				1: {ID: 1, OrderID: 100, ProductID: 200, Quantity: 2},
			},
			userID:      1,
			orderID:     100,
			productID:   200,
			actorUserID: 1,
			wantErr:     false,
			checkRemains: func(t *testing.T, repo *MockOrder) {
				for _, it := range repo.orderItems {
					if it.OrderID == 100 && it.ProductID == 200 {
						t.Error("item still exists after deletion")
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
			name: "item not found",
			seedOrders: map[int]orderRow{
				100: {ID: 100, UserID: 1, StoreID: 10, StoreName: "Store", CreatedAt: now, UpdatedAt: now},
			},
			seedItems:   map[int]orderItemRow{},
			userID:      1,
			orderID:     100,
			productID:   200,
			actorUserID: 1,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockOrder()
			repoMock.orders = tt.seedOrders
			repoMock.orderItems = tt.seedItems
			svc := service.NewServiceOrderItem(txMock, repoMock, repoMock)
			actor := policy.Actor{UserID: tt.actorUserID}

			err := svc.DeleteItem(context.Background(), actor, tt.orderID, tt.productID)

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
