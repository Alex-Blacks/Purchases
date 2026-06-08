package user_tests

import (
	"context"
	"errors"
	"testing"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/Alex-Blacks/Purchases/internal/policy"
	"github.com/Alex-Blacks/Purchases/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func testActor(userID int, role string) policy.Actor {
	actor := policy.Actor{UserID: userID, Role: policy.Role(role)}
	return actor
}

func hashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func TestUser_CreateUser(t *testing.T) {
	tests := []struct {
		name        string
		seedUsers   map[int]domain.User
		inputName   string
		inputEmail  string
		inputPass   string
		inputRole   string
		inputStatus string
		wantErr     bool
		wantErrIs   error
		checkTxFunc func(t *testing.T, tx *MockTx)
	}{
		{
			name:        "success",
			seedUsers:   map[int]domain.User{},
			inputName:   "John Doe",
			inputEmail:  "john@example.com",
			inputPass:   "password123",
			inputRole:   string(string(policy.RoleUser)),
			inputStatus: "active",
			wantErr:     false,
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
			name: "email already exists",
			seedUsers: map[int]domain.User{
				1: {ID: 1, Name: "Existing", Email: "existing@example.com", PasswordHash: hashPassword("pass"), Role: string(string(policy.RoleUser)), Status: "active"},
			},
			inputName:   "New User",
			inputEmail:  "existing@example.com",
			inputPass:   "password123",
			inputRole:   string(string(policy.RoleUser)),
			inputStatus: "active",
			wantErr:     true,
			wantErrIs:   domain.ErrEmailConflict,
			checkTxFunc: func(t *testing.T, tx *MockTx) {
				if tx.committed {
					t.Error("transaction committed despite error")
				}
				// Транзакция не начиналась, так как ошибка до неё
				if tx.rolledBack {
					t.Error("transaction rolled back but no transaction started")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockUser()
			for id, user := range tt.seedUsers {
				repoMock.users[id] = user
				if id >= repoMock.nextID {
					repoMock.nextID = id + 1
				}
			}
			svc := service.NewServiceUser(txMock, repoMock)

			user, err := svc.CreateUser(context.Background(), tt.inputName, tt.inputPass, tt.inputEmail, tt.inputRole, tt.inputStatus)

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
				if user.Name != tt.inputName {
					t.Errorf("expected name %q, got %q", tt.inputName, user.Name)
				}
				if user.Email != tt.inputEmail {
					t.Errorf("expected email %q, got %q", tt.inputEmail, user.Email)
				}
				if user.Role != tt.inputRole {
					t.Errorf("expected role %q, got %q", tt.inputRole, user.Role)
				}
				if user.Status != tt.inputStatus {
					t.Errorf("expected status %q, got %q", tt.inputStatus, user.Status)
				}
				if user.ID == 0 {
					t.Error("user ID not assigned")
				}
				if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(tt.inputPass)); err != nil {
					t.Error("password not properly hashed")
				}
			}
			if tt.checkTxFunc != nil {
				tt.checkTxFunc(t, txMock)
			}
		})
	}
}

func TestUser_GetUserByID(t *testing.T) {
	hashedPass := hashPassword("password123")
	tests := []struct {
		name        string
		seedUsers   map[int]domain.User
		actorUserID int
		actorRole   string
		userID      int
		wantErr     bool
		wantErrIs   error
		wantUser    domain.User
	}{
		{
			name: "success - own user",
			seedUsers: map[int]domain.User{
				1: {ID: 1, Name: "John", Email: "john@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
			},
			actorUserID: 1,
			userID:      1,
			wantErr:     false,
			wantUser:    domain.User{ID: 1, Name: "John", Email: "john@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
		},
		{
			name: "success - admin accesses other user",
			seedUsers: map[int]domain.User{
				1: {ID: 1, Name: "Admin", Email: "admin@example.com", PasswordHash: hashedPass, Role: string(policy.RoleAdmin), Status: "active"},
				2: {ID: 2, Name: "John", Email: "john@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
			},
			actorUserID: 1,
			actorRole:   string(policy.RoleAdmin),
			userID:      2,
			wantErr:     false,
			wantUser:    domain.User{ID: 2, Name: "John", Email: "john@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
		},
		{
			name: "forbidden - user accesses another user",
			seedUsers: map[int]domain.User{
				1: {ID: 1, Name: "User1", Email: "user1@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
				2: {ID: 2, Name: "User2", Email: "user2@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
			},
			actorUserID: 1,
			userID:      2,
			wantErr:     true,
			wantErrIs:   policy.ErrForbidden,
		},
		{
			name:        "user not found",
			seedUsers:   map[int]domain.User{},
			actorUserID: 1,
			userID:      999,
			wantErr:     true,
			wantErrIs:   domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockUser()
			for id, user := range tt.seedUsers {
				repoMock.users[id] = user
				if id >= repoMock.nextID {
					repoMock.nextID = id + 1
				}
			}
			svc := service.NewServiceUser(txMock, repoMock)

			user, err := svc.GetUserByID(context.Background(), testActor(tt.actorUserID, tt.actorRole), tt.userID)

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
				if user.ID != tt.wantUser.ID {
					t.Errorf("expected ID %d, got %d", tt.wantUser.ID, user.ID)
				}
				if user.Name != tt.wantUser.Name {
					t.Errorf("expected name %q, got %q", tt.wantUser.Name, user.Name)
				}
				if user.Email != tt.wantUser.Email {
					t.Errorf("expected email %q, got %q", tt.wantUser.Email, user.Email)
				}
			}
		})
	}
}

func TestUser_GetUserByEmail(t *testing.T) {
	hashedPass := hashPassword("password123")
	tests := []struct {
		name      string
		seedUsers map[int]domain.User
		email     string
		wantErr   bool
		wantErrIs error
		wantUser  domain.User
	}{
		{
			name: "success",
			seedUsers: map[int]domain.User{
				1: {ID: 1, Name: "John", Email: "john@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
			},
			email:    "john@example.com",
			wantErr:  false,
			wantUser: domain.User{ID: 1, Name: "John", Email: "john@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
		},
		{
			name:      "not found",
			seedUsers: map[int]domain.User{},
			email:     "nonexistent@example.com",
			wantErr:   true,
			wantErrIs: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockUser()
			for id, user := range tt.seedUsers {
				repoMock.users[id] = user
			}
			svc := service.NewServiceUser(txMock, repoMock)

			user, err := svc.GetUserByEmail(context.Background(), tt.email)

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
				if user != tt.wantUser {
					t.Errorf("expected %+v, got %+v", tt.wantUser, user)
				}
			}
		})
	}
}

func TestUser_UpdateUser(t *testing.T) {
	hashedPass := hashPassword("password123")
	tests := []struct {
		name         string
		seedUsers    map[int]domain.User
		actorUserID  int
		actorRole    string
		userID       int
		update       domain.UpdateUser
		wantErr      bool
		wantErrIs    error
		checkTxFunc  func(t *testing.T, tx *MockTx)
		checkUpdated func(t *testing.T, user domain.User)
	}{
		{
			name: "success - update own name",
			seedUsers: map[int]domain.User{
				1: {ID: 1, Name: "John", Email: "john@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
			},
			actorUserID: 1,
			userID:      1,
			update:      domain.UpdateUser{Name: ptr("John Updated")},
			wantErr:     false,
			checkTxFunc: func(t *testing.T, tx *MockTx) {
				if !tx.committed {
					t.Error("transaction not committed")
				}
				if tx.rolledBack {
					t.Error("transaction should not be rolled back")
				}
			},
			checkUpdated: func(t *testing.T, user domain.User) {
				if user.Name != "John Updated" {
					t.Errorf("expected name %q, got %q", "John Updated", user.Name)
				}
			},
		},
		{
			name: "success - admin updates role",
			seedUsers: map[int]domain.User{
				1: {ID: 1, Name: "Admin", Email: "admin@example.com", PasswordHash: hashedPass, Role: string(policy.RoleAdmin), Status: "active"},
				2: {ID: 2, Name: "John", Email: "john@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
			},
			actorUserID: 1,
			actorRole:   string(policy.RoleAdmin),
			userID:      2,
			update:      domain.UpdateUser{Role: ptr(string(policy.RoleAdmin))},
			wantErr:     false,
			checkUpdated: func(t *testing.T, user domain.User) {
				if user.Role != string(policy.RoleAdmin) {
					t.Errorf("expected role %q, got %q", policy.RoleAdmin, user.Role)
				}
			},
		},
		{
			name: "success - update password",
			seedUsers: map[int]domain.User{
				1: {ID: 1, Name: "John", Email: "john@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
			},
			actorUserID: 1,
			userID:      1,
			update:      domain.UpdateUser{Password: ptr("newpassword")},
			wantErr:     false,
			checkUpdated: func(t *testing.T, user domain.User) {
				if user.PasswordHash == hashedPass {
					t.Error("password was not updated")
				}
				if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("newpassword")); err != nil {
					t.Error("new password not properly hashed")
				}
			},
		},
		{
			name: "success - update email",
			seedUsers: map[int]domain.User{
				1: {ID: 1, Name: "John", Email: "john@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
			},
			actorUserID: 1,
			userID:      1,
			update:      domain.UpdateUser{Email: ptr("john.new@example.com")},
			wantErr:     false,
			checkUpdated: func(t *testing.T, user domain.User) {
				if user.Email != "john.new@example.com" {
					t.Errorf("expected email %q, got %q", "john.new@example.com", user.Email)
				}
			},
		},
		{
			name: "forbidden - user updates role",
			seedUsers: map[int]domain.User{
				1: {ID: 1, Name: "John", Email: "john@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
			},
			actorUserID: 1,
			userID:      1,
			update:      domain.UpdateUser{Role: ptr(string(policy.RoleAdmin))},
			wantErr:     true,
			wantErrIs:   policy.ErrForbidden,
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
			name: "no fields to update",
			seedUsers: map[int]domain.User{
				1: {ID: 1, Name: "John", Email: "john@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
			},
			actorUserID: 1,
			userID:      1,
			update:      domain.UpdateUser{},
			wantErr:     true,
			wantErrIs:   domain.ErrNoFieldsToUpdate,
		},
		{
			name: "email conflict",
			seedUsers: map[int]domain.User{
				1: {ID: 1, Name: "John", Email: "john@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
				2: {ID: 2, Name: "Jane", Email: "jane@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
			},
			actorUserID: 1,
			userID:      1,
			update:      domain.UpdateUser{Email: ptr("jane@example.com")},
			wantErr:     true,
			wantErrIs:   domain.ErrConflict,
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
			name: "user not found",
			seedUsers: map[int]domain.User{
				1: {ID: 1, Name: "John", Email: "john@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
			},
			actorUserID: 1,
			userID:      999,
			update:      domain.UpdateUser{Name: ptr("Updated")},
			wantErr:     true,
			wantErrIs:   domain.ErrNotFound,
			checkTxFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockUser()
			for id, user := range tt.seedUsers {
				repoMock.users[id] = user
				if id >= repoMock.nextID {
					repoMock.nextID = id + 1
				}
			}
			svc := service.NewServiceUser(txMock, repoMock)

			user, err := svc.UpdateUser(context.Background(), testActor(tt.actorUserID, tt.actorRole), tt.userID, tt.update)

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
				if tt.checkUpdated != nil {
					tt.checkUpdated(t, user)
				}
			}
			if tt.checkTxFunc != nil {
				tt.checkTxFunc(t, txMock)
			}
		})
	}
}

func TestUser_DeleteUser(t *testing.T) {
	hashedPass := hashPassword("password123")
	tests := []struct {
		name         string
		seedUsers    map[int]domain.User
		actorUserID  int
		actorRole    string
		userID       int
		wantErr      bool
		wantErrIs    error
		checkRemains func(t *testing.T, repo *MockUser)
		checkTxFunc  func(t *testing.T, tx *MockTx)
	}{
		{
			name: "success - delete own user",
			seedUsers: map[int]domain.User{
				1: {ID: 1, Name: "John", Email: "john@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
			},
			actorUserID: 1,
			userID:      1,
			wantErr:     false,
			checkRemains: func(t *testing.T, repo *MockUser) {
				if _, ok := repo.users[1]; ok {
					t.Error("user still exists after deletion")
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
			name: "success - admin deletes other user",
			seedUsers: map[int]domain.User{
				1: {ID: 1, Name: "Admin", Email: "admin@example.com", PasswordHash: hashedPass, Role: string(policy.RoleAdmin), Status: "active"},
				2: {ID: 2, Name: "John", Email: "john@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
			},
			actorUserID: 1,
			actorRole:   string(policy.RoleAdmin),
			userID:      2,
			wantErr:     false,
			checkRemains: func(t *testing.T, repo *MockUser) {
				if _, ok := repo.users[2]; ok {
					t.Error("user still exists after deletion")
				}
				if _, ok := repo.users[1]; !ok {
					t.Error("admin user should not be deleted")
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
			name: "forbidden - user deletes other user",
			seedUsers: map[int]domain.User{
				1: {ID: 1, Name: "User1", Email: "user1@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
				2: {ID: 2, Name: "User2", Email: "user2@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
			},
			actorUserID: 1,
			userID:      2,
			wantErr:     true,
			wantErrIs:   policy.ErrForbidden,
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
			name:        "user not found",
			seedUsers:   map[int]domain.User{},
			actorUserID: 1,
			userID:      999,
			wantErr:     true,
			wantErrIs:   domain.ErrNotFound,
			checkTxFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockUser()
			for id, user := range tt.seedUsers {
				repoMock.users[id] = user
				if id >= repoMock.nextID {
					repoMock.nextID = id + 1
				}
			}
			svc := service.NewServiceUser(txMock, repoMock)

			err := svc.DeleteUser(context.Background(), testActor(tt.actorUserID, tt.actorRole), tt.userID)

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

func TestUser_ListUsers(t *testing.T) {
	hashedPass := hashPassword("password123")
	tests := []struct {
		name        string
		seedUsers   map[int]domain.User
		actorUserID int
		actorRole   string
		wantErr     bool
		wantErrIs   error
		wantCount   int
	}{
		{
			name: "success - admin lists all users",
			seedUsers: map[int]domain.User{
				1: {ID: 1, Name: "Admin", Email: "admin@example.com", PasswordHash: hashedPass, Role: string(policy.RoleAdmin), Status: "active"},
				2: {ID: 2, Name: "John", Email: "john@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
				3: {ID: 3, Name: "Jane", Email: "jane@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "inactive"},
			},
			actorUserID: 1,
			actorRole:   string(policy.RoleAdmin),
			wantErr:     false,
			wantCount:   3,
		},
		{
			name: "forbidden - user cannot list users",
			seedUsers: map[int]domain.User{
				1: {ID: 1, Name: "John", Email: "john@example.com", PasswordHash: hashedPass, Role: string(policy.RoleUser), Status: "active"},
			},
			actorUserID: 1,
			wantErr:     true,
			wantErrIs:   policy.ErrForbidden,
		},
		{
			name:        "empty list - admin with no users",
			seedUsers:   map[int]domain.User{},
			actorUserID: 1,
			actorRole:   string(policy.RoleAdmin),
			wantErr:     false,
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockUser()
			for id, user := range tt.seedUsers {
				repoMock.users[id] = user
			}
			svc := service.NewServiceUser(txMock, repoMock)

			users, err := svc.ListUsers(context.Background(), testActor(tt.actorUserID, tt.actorRole))

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
				if len(users) != tt.wantCount {
					t.Errorf("expected %d users, got %d", tt.wantCount, len(users))
				}
				// Проверяем, что список отсортирован по ID
				for i := 1; i < len(users); i++ {
					if users[i].ID < users[i-1].ID {
						t.Error("users not sorted by ID")
					}
				}
			}
		})
	}
}

func TestUser_CheckPassword(t *testing.T) {
	hashedPass := hashPassword("correctpassword")
	tests := []struct {
		name      string
		user      domain.User
		password  string
		wantErr   bool
		wantErrIs error
	}{
		{
			name:     "success",
			user:     domain.User{PasswordHash: hashedPass},
			password: "correctpassword",
			wantErr:  false,
		},
		{
			name:      "wrong password",
			user:      domain.User{PasswordHash: hashedPass},
			password:  "wrongpassword",
			wantErr:   true,
			wantErrIs: bcrypt.ErrMismatchedHashAndPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockUser()
			svc := service.NewServiceUser(txMock, repoMock)

			err := svc.CheckPassword(tt.user, tt.password)

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
		})
	}
}

func TestUser_GeneratePassword(t *testing.T) {
	txMock := NewMockTx()
	repoMock := NewMockUser()
	svc := service.NewServiceUser(txMock, repoMock)

	password := "testpassword123"
	hash, err := svc.GeneratePassword(password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "" {
		t.Error("password hash is empty")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		t.Error("generated hash does not match original password")
	}
}

func ptr(s string) *string {
	return &s
}

func TestAuthService_Login(t *testing.T) {
	secret := "test-secret-key"

	tests := []struct {
		name          string
		seedUsers     map[int]domain.User
		email         string
		password      string
		wantErr       bool
		wantErrIs     error
		validateToken func(t *testing.T, token string)
	}{
		{
			name: "success - valid credentials",
			seedUsers: map[int]domain.User{
				1: {
					ID:           1,
					Name:         "John Doe",
					Email:        "john@example.com",
					PasswordHash: hashPassword("password123"),
					Role:         string(policy.RoleUser),
					Status:       "active",
				},
			},
			email:    "john@example.com",
			password: "password123",
			wantErr:  false,
			validateToken: func(t *testing.T, token string) {
				claims, err := parseToken(token, secret)
				if err != nil {
					t.Fatalf("failed to parse token: %v", err)
				}
				sub, ok := claims["sub"].(float64)
				if !ok || int(sub) != 1 {
					t.Errorf("expected sub=1, got %v", claims["sub"])
				}
				role, ok := claims["role"].(string)
				if !ok || role != string(policy.RoleUser) {
					t.Errorf("expected role=%s, got %v", policy.RoleUser, claims["role"])
				}
			},
		},
		{
			name: "success - admin user",
			seedUsers: map[int]domain.User{
				2: {
					ID:           2,
					Name:         "Admin User",
					Email:        "admin@example.com",
					PasswordHash: hashPassword("admin123"),
					Role:         string(policy.RoleAdmin),
					Status:       "active",
				},
			},
			email:    "admin@example.com",
			password: "admin123",
			wantErr:  false,
			validateToken: func(t *testing.T, token string) {
				claims, err := parseToken(token, secret)
				if err != nil {
					t.Fatalf("failed to parse token: %v", err)
				}
				role, ok := claims["role"].(string)
				if !ok || role != string(policy.RoleAdmin) {
					t.Errorf("expected role=%s, got %v", policy.RoleAdmin, claims["role"])
				}
			},
		},
		{
			name:      "error - user not found",
			seedUsers: map[int]domain.User{},
			email:     "nonexistent@example.com",
			password:  "password123",
			wantErr:   true,
			wantErrIs: domain.ErrNotFound,
		},
		{
			name: "error - user status blocked",
			seedUsers: map[int]domain.User{
				3: {
					ID:           3,
					Name:         "Blocked User",
					Email:        "blocked@example.com",
					PasswordHash: hashPassword("blocked123"),
					Role:         string(policy.RoleUser),
					Status:       "blocked",
				},
			},
			email:     "blocked@example.com",
			password:  "blocked123",
			wantErr:   true,
			wantErrIs: domain.ErrStatusBlocked,
		},
		{
			name: "error - invalid password",
			seedUsers: map[int]domain.User{
				4: {
					ID:           4,
					Name:         "Test User",
					Email:        "test@example.com",
					PasswordHash: hashPassword("correctpassword"),
					Role:         string(policy.RoleUser),
					Status:       "active",
				},
			},
			email:     "test@example.com",
			password:  "wrongpassword",
			wantErr:   true,
			wantErrIs: domain.ErrIncorrectPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMock := NewMockTx()
			repoMock := NewMockUser()
			for id, user := range tt.seedUsers {
				repoMock.users[id] = user
				if id >= repoMock.nextID {
					repoMock.nextID = id + 1
				}
			}
			userSvc := service.NewServiceUser(txMock, repoMock)
			authSvc := service.NewAuthService(userSvc, secret)

			token, err := authSvc.Login(context.Background(), tt.email, tt.password)

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
				if token == "" {
					t.Error("token is empty")
				}
				if tt.validateToken != nil {
					tt.validateToken(t, token)
				}
			}
		})
	}
}

func parseToken(tokenString, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
