package order_tests

import (
	"context"
	"errors"
	"sort"
	"time"

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

type orderRow struct {
	ID        int
	UserID    int
	StoreID   int
	StoreName string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type orderItemRow struct {
	ID        int
	OrderID   int
	ProductID int
	Quantity  int
}

type MockOrder struct {
	orders      map[int]orderRow
	orderItems  map[int]orderItemRow
	nextOrderID int
	nextItemID  int
	stores      map[int]bool
	products    map[int]string
}

func NewMockOrder() *MockOrder {
	return &MockOrder{
		orders:      make(map[int]orderRow),
		orderItems:  make(map[int]orderItemRow),
		nextOrderID: 1,
		nextItemID:  1,
		stores:      make(map[int]bool),
		products:    make(map[int]string),
	}
}

func (m *MockOrder) AddStore(id int) {
	m.stores[id] = true
}

func (m *MockOrder) AddProduct(id int, title string) {
	m.products[id] = title
}

func (m *MockOrder) getOrderWithItems(orderID int) (domain.OrderWithItemDetails, error) {
	order, ok := m.orders[orderID]
	if !ok {
		return domain.OrderWithItemDetails{}, domain.ErrNotFound
	}
	var items []domain.OrderItemDetails
	for _, it := range m.orderItems {
		if it.OrderID == orderID {
			title, ok := m.products[it.ProductID]
			if !ok {
				title = "unknown_product"
			}
			items = append(items, domain.OrderItemDetails{
				ID:        it.ID,
				ProductID: it.ProductID,
				Title:     title,
				Quantity:  it.Quantity,
			})
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })

	return domain.OrderWithItemDetails{
		Order: domain.OrderDetails{
			ID:         order.ID,
			UserID:     order.UserID,
			Store:      order.StoreName,
			ItemsCount: len(items),
			CreatedAt:  order.CreatedAt,
			UpdatedAt:  order.UpdatedAt,
		},
		Items: items,
	}, nil
}

func (m *MockOrder) CreateOrder(ctx context.Context, q domain.Querier, userID, storeID int) (domain.OrderWithItemDetails, error) {
	if !m.stores[storeID] {
		return domain.OrderWithItemDetails{}, domain.ErrConflict
	}
	id := m.nextOrderID
	m.nextOrderID++
	now := time.Now().UTC()
	order := orderRow{
		ID:        id,
		UserID:    userID,
		StoreID:   storeID,
		StoreName: "Store_" + string(rune(storeID)),
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.orders[id] = order
	return m.getOrderWithItems(id)
}

func (m *MockOrder) GetOrder(ctx context.Context, q domain.Querier, userID, orderID int) (domain.OrderWithItemDetails, error) {
	order, err := m.getOrderWithItems(orderID)
	if err != nil {
		return domain.OrderWithItemDetails{}, err
	}
	if order.Order.UserID != userID {
		return domain.OrderWithItemDetails{}, domain.ErrNotFound
	}
	return order, nil
}

func (m *MockOrder) DeleteOrder(ctx context.Context, q domain.Querier, userID, orderID int) error {
	order, ok := m.orders[orderID]
	if !ok {
		return domain.ErrNotFound
	}
	if order.UserID != userID {
		return domain.ErrNotFound
	}
	for id, it := range m.orderItems {
		if it.OrderID == orderID {
			delete(m.orderItems, id)
		}
	}
	delete(m.orders, orderID)
	return nil
}

func (m *MockOrder) ListOrders(ctx context.Context, q domain.Querier, userID int) ([]domain.OrderDetails, error) {
	var result []domain.OrderDetails
	for _, order := range m.orders {
		if order.UserID != userID {
			continue
		}
		itemsCount := 0
		for _, it := range m.orderItems {
			if it.OrderID == order.ID {
				itemsCount++
			}
		}
		result = append(result, domain.OrderDetails{
			ID:         order.ID,
			UserID:     order.UserID,
			Store:      order.StoreName,
			ItemsCount: itemsCount,
			CreatedAt:  order.CreatedAt,
			UpdatedAt:  order.UpdatedAt,
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func (m *MockOrder) orderExists(orderID int) bool {
	_, ok := m.orders[orderID]
	return ok
}

func (m *MockOrder) AddItem(ctx context.Context, q domain.Querier, orderID, productID, quantity int) (domain.OrderItemDetails, error) {
	if !m.orderExists(orderID) {
		return domain.OrderItemDetails{}, domain.ErrNotFound
	}
	if _, ok := m.products[productID]; !ok {
		return domain.OrderItemDetails{}, domain.ErrConflict
	}
	for _, it := range m.orderItems {
		if it.OrderID == orderID && it.ProductID == productID {
			return domain.OrderItemDetails{}, domain.ErrAlreadyExists
		}
	}
	id := m.nextItemID
	m.nextItemID++
	item := orderItemRow{
		ID:        id,
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  quantity,
	}
	m.orderItems[id] = item
	title, _ := m.products[productID]
	return domain.OrderItemDetails{
		ID:        id,
		ProductID: productID,
		Title:     title,
		Quantity:  quantity,
	}, nil
}

func (m *MockOrder) UpdateItem(ctx context.Context, q domain.Querier, orderID, productID, quantity int) (domain.OrderItemDetails, error) {
	var foundID int
	for id, it := range m.orderItems {
		if it.OrderID == orderID && it.ProductID == productID {
			foundID = id
			break
		}
	}
	if foundID == 0 {
		return domain.OrderItemDetails{}, domain.ErrNotFound
	}
	item := m.orderItems[foundID]
	item.Quantity += quantity
	m.orderItems[foundID] = item
	title, _ := m.products[productID]
	return domain.OrderItemDetails{
		ID:        foundID,
		ProductID: productID,
		Title:     title,
		Quantity:  item.Quantity,
	}, nil
}

func (m *MockOrder) DeleteItem(ctx context.Context, q domain.Querier, orderID, productID int) error {
	var foundID int
	for id, it := range m.orderItems {
		if it.OrderID == orderID && it.ProductID == productID {
			foundID = id
			break
		}
	}
	if foundID == 0 {
		return domain.ErrNotFound
	}
	delete(m.orderItems, foundID)
	return nil
}
