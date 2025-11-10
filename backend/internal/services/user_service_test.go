package services

import (
	"context"
	"sync"
	"testing"
)

type memoryUserRepo struct {
	mu    sync.Mutex
	users map[string]User
}

func newMemoryUserRepo() *memoryUserRepo {
	return &memoryUserRepo{users: make(map[string]User)}
}

func (m *memoryUserRepo) FindByEmail(_ context.Context, email string) (User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, ok := m.users[email]
	if !ok {
		return User{}, ErrNotFound
	}
	return user, nil
}

func (m *memoryUserRepo) Insert(_ context.Context, user User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.users[user.Email] = user
	return nil
}

func (m *memoryUserRepo) List(_ context.Context) ([]User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]User, 0, len(m.users))
	for _, u := range m.users {
		result = append(result, u)
	}
	return result, nil
}

func (m *memoryUserRepo) Clear(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.users = make(map[string]User)
	return nil
}

// TestUserServiceRegisterStoresNormalizedUsers ensures Register persists sanitized data.
func TestUserServiceRegisterStoresNormalizedUsers(t *testing.T) {
	ctx := context.Background()
	repo := newMemoryUserRepo()
	service := NewUserService(repo)

	err := service.Register(ctx, User{Email: " User@Example.com ", Password: " secret "})
	if err != nil {
		t.Fatalf("expected register to succeed, got %v", err)
	}

	stored, err := repo.FindByEmail(ctx, "user@example.com")
	if err != nil {
		t.Fatalf("expected user to be stored, got %v", err)
	}
	if stored.Email != "user@example.com" {
		t.Errorf("expected normalized email, got %q", stored.Email)
	}
	if stored.Password != "secret" {
		t.Errorf("expected trimmed password, got %q", stored.Password)
	}
}

// TestUserServiceRegisterRejectsDuplicates verifies duplicate emails fail.
func TestUserServiceRegisterRejectsDuplicates(t *testing.T) {
	ctx := context.Background()
	service := NewUserService(newMemoryUserRepo())

	if err := service.Register(ctx, User{Email: "user@example.com", Password: "secret"}); err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	err := service.Register(ctx, User{Email: "user@example.com", Password: "secret"})
	if err != ErrUserAlreadyExists {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
}

// TestUserServiceLoginValidatesCredentials exercises success and failure cases.
func TestUserServiceLoginValidatesCredentials(t *testing.T) {
	ctx := context.Background()
	service := NewUserService(newMemoryUserRepo())

	if err := service.Register(ctx, User{Email: "user@example.com", Password: "secret"}); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if err := service.Login(ctx, " User@Example.com ", " secret "); err != nil {
		t.Fatalf("expected login to succeed, got %v", err)
	}

	if err := service.Login(ctx, "user@example.com", "wrong"); err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials for wrong password, got %v", err)
	}

	if err := service.Login(ctx, "", "secret"); err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials for missing email, got %v", err)
	}
}

// TestUserServiceListAndClear ensures List returns public data and Clear removes users.
func TestUserServiceListAndClear(t *testing.T) {
	ctx := context.Background()
	repo := newMemoryUserRepo()
	service := NewUserService(repo)

	users := []User{
		{Email: "alice@example.com", Password: "alice"},
		{Email: "bob@example.com", Password: "bob"},
	}
	for _, u := range users {
		if err := service.Register(ctx, u); err != nil {
			t.Fatalf("register failed: %v", err)
		}
	}

	public, err := service.List(ctx)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(public) != 2 {
		t.Fatalf("expected 2 users, got %d", len(public))
	}
	for _, u := range public {
		if u.Email == "" {
			t.Errorf("expected email to be present, got empty string")
		}
	}

	if err := service.Clear(ctx); err != nil {
		t.Fatalf("clear failed: %v", err)
	}

	remaining, err := service.List(ctx)
	if err != nil {
		t.Fatalf("list after clear failed: %v", err)
	}
	if len(remaining) != 0 {
		t.Fatalf("expected 0 users after clear, got %d", len(remaining))
	}
}
