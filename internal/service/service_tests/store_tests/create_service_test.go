package store_tests

import (
	"context"
	"errors"
	"testing"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

func TestStore_CreateStore(t *testing.T) {
	tests := []struct {
		name     string
		seed     map[int]domain.Store
		wantErr  bool
		wantDate domain.Store
	}{
		{
			name:    "success",
			seed:    map[int]domain.Store{},
			wantErr: false,
			wantDate: domain.Store{
				ID:   1,
				Name: "Test1",
			},
		},
		{
			name: "already exists",
			seed: map[int]domain.Store{
				1: {
					ID:   1,
					Name: "Test1",
				},
			},
			wantErr:  true,
			wantDate: domain.Store{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := NewTestServiceStore(tt.seed)

			store, err := svc.CreateStore(context.Background(), "Test1")
			if err != nil {
				if tt.wantErr {
					if errors.Is(err, domain.ErrAlreadyExists) {
						return
					}
					t.Fatalf("expected error: %v, got: %v", err, domain.ErrAlreadyExists)
				}
				t.Fatal("unexpected error")
				return
			}

			if tt.wantDate.ID != store.ID || tt.wantDate.Name != store.Name {
				t.Fatalf("expected ID: %d, name: %s / got id: %d, name: %s", tt.wantDate.ID, tt.wantDate.Name, store.ID, store.Name)
			}
		})
	}
}
