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

func TestOrder_AddItem(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name         string
		seedOrders   map[int]orderRow
		seedItems    map[int]orderItemRow
		seedProducts map[int]string
		actorUserID  int
		orderID      int
		productID    int
		quantity     int
		wantErr      bool
		wantErrIs    error
		checkItem    func(t *testing.T, item domain.OrderItemDetails)
		checkTxFunc  func(t *testing.T, tx *MockTx)
	}{
		{
			name: "success - add new item",
			seedOrders: map[int]orderRow{
				100: {ID: 100, UserID: 1, StoreID: 10, StoreName: "Store", CreatedAt: now, UpdatedAt: now},
			},
			seedProducts: map[int]string{200: "Apple"},
			actorUserID:  1,
			orderID:      100,
			productID:    200,
			quantity:     3,
			wantErr:      false,
			checkItem: func(t *testing.T, item domain.OrderItemDetails) {
				if item.ProductID != 200 {
					t.Errorf("expected ProductID 200, got %d", item.ProductID)
				}
				if item.Quantity != 3 {
					t.Errorf("expected Quantity 3, got %d", item.Quantity)
				}
				if item.Title != "Apple" {
					t.Errorf("expected Title Apple, got %s", item.Title)
				}
				if item.ID == 0 {
					t.Error("item ID not assigned")
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
			name: "success - item already exists, should update quantity",
			seedOrders: map[int]orderRow{
				101: {ID: 101, UserID: 1, StoreID: 10, StoreName: "Store", CreatedAt: now, UpdatedAt: now},
			},
			seedItems: map[int]orderItemRow{
				1000: {ID: 1000, OrderID: 101, ProductID: 201, Quantity: 2},
			},
			seedProducts: map[int]string{201: "Banana"},
			actorUserID:  1,
			orderID:      101,
			productID:    201,
			quantity:     4,
			wantErr:      false,
			checkItem: func(t *testing.T, item domain.OrderItemDetails) {
				// В реальном сервисе AddItem при существующем элементе вызывает UpdateItem с суммированием
				// Ожидаем, что quantity станет 2+4 = 6
				if item.Quantity != 6 {
					t.Errorf("expected Quantity 6 (2+4), got %d", item.Quantity)
				}
				if item.ProductID != 201 {
					t.Errorf("expected ProductID 201, got %d", item.ProductID)
				}
				if item.Title != "Banana" {
					t.Errorf("expected Title Banana, got %s", item.Title)
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
			actorUserID: 1,
			orderID:     999,
			productID:   200,
			quantity:    1,
			wantErr:     true,
			wantErrIs:   domain.ErrNotFound,
			checkTxFunc: func(t *testing.T, tx *MockTx) {
				// Ошибка происходит до начала транзакции (в GetAccessibleOrder)
				// Транзакция не создаётся, поэтому флаги не меняются
				if tx.committed {
					t.Error("transaction committed despite error")
				}
				if tx.rolledBack {
					t.Error("transaction rolled back but no transaction started")
				}
			},
		},
		{
			name: "order belongs to another user",
			seedOrders: map[int]orderRow{
				102: {ID: 102, UserID: 2, StoreID: 10, StoreName: "Store", CreatedAt: now, UpdatedAt: now},
			},
			actorUserID: 1,
			orderID:     102,
			productID:   200,
			quantity:    1,
			wantErr:     true,
			wantErrIs:   domain.ErrNotFound,
			checkTxFunc: func(t *testing.T, tx *MockTx) {
				if tx.committed {
					t.Error("transaction committed despite error")
				}
				if tx.rolledBack {
					t.Error("transaction rolled back but no transaction started")
				}
			},
		},
		{
			name: "product not found (conflict)",
			seedOrders: map[int]orderRow{
				103: {ID: 103, UserID: 1, StoreID: 10, StoreName: "Store", CreatedAt: now, UpdatedAt: now},
			},
			seedProducts: map[int]string{}, // продукт не добавлен
			actorUserID:  1,
			orderID:      103,
			productID:    999,
			quantity:     1,
			wantErr:      true,
			wantErrIs:    domain.ErrConflict, // внешний ключ нарушен
			checkTxFunc: func(t *testing.T, tx *MockTx) {
				// Ошибка происходит внутри транзакции, должен быть rollback
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
			// заполняем заказы
			for id, row := range tt.seedOrders {
				repoMock.orders[id] = row
				if id >= repoMock.nextOrderID {
					repoMock.nextOrderID = id + 1
				}
			}
			// заполняем позиции
			for id, row := range tt.seedItems {
				repoMock.orderItems[id] = row
				if id >= repoMock.nextItemID {
					repoMock.nextItemID = id + 1
				}
			}
			// заполняем продукты
			for id, title := range tt.seedProducts {
				repoMock.AddProduct(id, title)
			}
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
				if item.Quantity != 9 {
					t.Errorf("expected quantity 9, got %d", item.Quantity)
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
