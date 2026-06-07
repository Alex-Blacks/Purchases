package product_tests

import (
	"context"
	"errors"
	"fmt"
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

type ProductInput struct {
	ID         int
	Title      string
	Unit       string
	CategoryID int
}

type AliasInput struct {
	ID        int
	ProductID int
	Alias     string
}

type MockProduct struct {
	products      map[int]ProductInput
	aliases       map[int]AliasInput
	nextProductID int
	nextAliasID   int
}

func NewMockProduct() *MockProduct {
	return &MockProduct{
		products:      make(map[int]ProductInput),
		aliases:       make(map[int]AliasInput),
		nextProductID: 1,
		nextAliasID:   1,
	}
}

func (m *MockProduct) getProductDetails(id int) (domain.ProductDetails, error) {
	p, ok := m.products[id]
	if !ok {
		return domain.ProductDetails{}, domain.ErrNotFound
	}
	return domain.ProductDetails{
		ID:       p.ID,
		Title:    p.Title,
		Unit:     p.Unit,
		Category: fmt.Sprintf("Category_%d", p.CategoryID),
	}, nil
}

func (m *MockProduct) CreateProduct(ctx context.Context, q domain.Querier, title, unit string, categoryID int) (domain.ProductDetails, error) {
	if _, ok := m.products[categoryID]; !ok {
		return domain.ProductDetails{}, domain.ErrConflict
	}
	for _, product := range m.products {
		if product.Title == title {
			return domain.ProductDetails{}, domain.ErrAlreadyExists
		}
	}
	id := m.nextProductID
	m.nextProductID++
	newProduct := ProductInput{ID: id, Title: title, Unit: unit, CategoryID: categoryID}
	m.products[id] = newProduct

	return m.getProductDetails(id)
}

func (m *MockProduct) GetProduct(ctx context.Context, q domain.Querier, id int) (domain.ProductDetails, error) {
	return m.getProductDetails(id)
}

func (m *MockProduct) DeleteProduct(ctx context.Context, q domain.Querier, id int) error {
	for _, a := range m.aliases {
		if a.ProductID == id {
			return domain.ErrConflict
		}

	}

	if _, ok := m.products[id]; !ok {
		return domain.ErrNotFound
	}

	delete(m.products, id)
	return nil
}

func (m *MockProduct) ListProducts(ctx context.Context, q domain.Querier) ([]domain.ProductDetails, error) {
	res := make([]domain.ProductDetails, 0, len(m.products))
	for id := range m.products {
		det, err := m.getProductDetails(id)
		if err != nil {
			continue
		}
		res = append(res, det)
	}

	sort.Slice(res, func(i, j int) bool { return res[i].ID < res[j].ID })
	return res, nil
}

func (m *MockProduct) CreateProductAlias(ctx context.Context, q domain.Querier, productID int, alias string) (domain.ProductAliasDetails, error) {
	if _, ok := m.products[productID]; !ok {
		return domain.ProductAliasDetails{}, domain.ErrConflict
	}
	for _, a := range m.aliases {
		if a.Alias == alias {
			return domain.ProductAliasDetails{}, domain.ErrAlreadyExists
		}
	}

	id := m.nextAliasID
	m.nextAliasID++
	newAlias := AliasInput{ID: id, ProductID: productID, Alias: alias}
	m.aliases[id] = newAlias

	return m.GetProductAlias(ctx, q, id)
}

func (m *MockProduct) GetProductAlias(ctx context.Context, q domain.Querier, id int) (domain.ProductAliasDetails, error) {
	a, ok := m.aliases[id]
	if !ok {
		return domain.ProductAliasDetails{}, domain.ErrNotFound
	}

	product, err := m.getProductDetails(a.ProductID)
	if err != nil {
		return domain.ProductAliasDetails{}, err
	}

	return domain.ProductAliasDetails{
		ID:      a.ID,
		Product: product.Title,
		Alias:   a.Alias,
	}, nil
}

func (m *MockProduct) DeleteProductAlias(ctx context.Context, q domain.Querier, id int) error {
	if _, ok := m.aliases[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.aliases, id)
	return nil
}

func (m *MockProduct) ListProductAliases(ctx context.Context, q domain.Querier, productID int) ([]domain.ProductAliasDetails, error) {
	var res []domain.ProductAliasDetails
	for _, a := range m.aliases {
		if a.ProductID == productID {
			det, err := m.GetProductAlias(ctx, q, a.ID)
			if err != nil {
				continue
			}
			res = append(res, det)
		}
	}

	sort.Slice(res, func(i, j int) bool { return res[i].ID < res[j].ID })
	return res, nil
}
func (m *MockProduct) DeleteAllProductAliases(ctx context.Context, q domain.Querier, productID int) error {
	if _, ok := m.products[productID]; !ok {
		return domain.ErrNotFound
	}
	deleted := false
	for id, a := range m.aliases {
		if a.ProductID == productID {
			delete(m.aliases, id)
			deleted = true
		}
	}

	if !deleted {
		return domain.ErrNotFound
	}
	return nil
}

func (m *MockProduct) FindProductByAlias(ctx context.Context, q domain.Querier, alias string) (string, error) {
	for _, a := range m.aliases {
		if a.Alias == alias {
			product, err := m.getProductDetails(a.ProductID)
			if err != nil {
				return "", err
			}
			return product.Title, nil
		}
	}
	return "", domain.ErrNotFound
}
