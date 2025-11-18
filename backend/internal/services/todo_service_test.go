package services

import (
	"context"
	"sort"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type memoryTodoRepo struct {
	todos map[primitive.ObjectID]Todo
}

func newMemoryTodoRepo() *memoryTodoRepo {
	return &memoryTodoRepo{todos: make(map[primitive.ObjectID]Todo)}
}

func (m *memoryTodoRepo) List(_ context.Context, email string) ([]Todo, error) {
	var result []Todo
	for _, todo := range m.todos {
		if email == "" || todo.Email == email {
			result = append(result, todo)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result, nil
}

func (m *memoryTodoRepo) Create(_ context.Context, todo Todo) (Todo, error) {
	if todo.ID.IsZero() {
		todo.ID = primitive.NewObjectID()
	}
	m.todos[todo.ID] = todo
	return todo, nil
}

func (m *memoryTodoRepo) Update(_ context.Context, id primitive.ObjectID, update TodoUpdate) (Todo, error) {
	todo, ok := m.todos[id]
	if !ok {
		return Todo{}, ErrNotFound
	}
	if update.Title != nil {
		todo.Title = *update.Title
	}
	if update.Completed != nil {
		todo.Completed = *update.Completed
	}
	m.todos[id] = todo
	return todo, nil
}

func (m *memoryTodoRepo) Delete(_ context.Context, id primitive.ObjectID) error {
	if _, ok := m.todos[id]; !ok {
		return ErrNotFound
	}
	delete(m.todos, id)
	return nil
}

func (m *memoryTodoRepo) Clear(_ context.Context, email string) error {
	for id, todo := range m.todos {
		if email == "" || todo.Email == email {
			delete(m.todos, id)
		}
	}
	return nil
}

func fixedNow() time.Time {
	return time.Date(2025, time.January, 1, 10, 0, 0, 0, time.UTC)
}

// TestTodoServiceCreateNormalizesInput asserts Create sanitizes fields and stores todos.
func TestTodoServiceCreateNormalizesInput(t *testing.T) {
	ctx := context.Background()
	repo := newMemoryTodoRepo()
	service := NewTodoService(repo, fixedNow)

	resp, err := service.Create(ctx, " User@Example.com ", " Primera tarea ")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if resp.Email != "user@example.com" {
		t.Errorf("expected normalized email, got %q", resp.Email)
	}
	if resp.Title != "Primera tarea" {
		t.Errorf("expected trimmed title, got %q", resp.Title)
	}
	if !resp.CreatedAt.Equal(fixedNow()) {
		t.Errorf("expected fixed timestamp, got %v", resp.CreatedAt)
	}
	if resp.Completed {
		t.Errorf("expected new todo to be incomplete")
	}
}

// TestTodoServiceListFiltersByEmail verifies List respects email filters.
func TestTodoServiceListFiltersByEmail(t *testing.T) {
	ctx := context.Background()
	repo := newMemoryTodoRepo()
	service := NewTodoService(repo, fixedNow)

	_, _ = service.Create(ctx, "alice@example.com", "Task A")
	_, _ = service.Create(ctx, "bob@example.com", "Task B")

	todos, err := service.List(ctx, "bob@example.com")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(todos) != 1 {
		t.Fatalf("expected 1 todo for bob, got %d", len(todos))
	}
	if todos[0].Email != "bob@example.com" {
		t.Errorf("expected todo for bob, got %q", todos[0].Email)
	}
}

// TestTodoServiceUpdateModifiesFields confirms Update validates data and persists changes.
func TestTodoServiceUpdateModifiesFields(t *testing.T) {
	ctx := context.Background()
	repo := newMemoryTodoRepo()
	service := NewTodoService(repo, fixedNow)

	created, err := service.Create(ctx, "alice@example.com", "Initial")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	newTitle := "Updated"
	done := true
	updated, err := service.Update(ctx, created.ID, TodoUpdate{
		Title:     &newTitle,
		Completed: &done,
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Title != "Updated" || !updated.Completed {
		t.Errorf("update did not apply correctly: %+v", updated)
	}

	_, err = service.Update(ctx, created.ID, TodoUpdate{})
	if err != ErrInvalidTodoInput {
		t.Fatalf("expected ErrInvalidTodoInput for empty update, got %v", err)
	}
}

// TestTodoServiceDeleteAndClear validates Delete and Clear flows.
func TestTodoServiceDeleteAndClear(t *testing.T) {
	ctx := context.Background()
	repo := newMemoryTodoRepo()
	service := NewTodoService(repo, fixedNow)

	first, _ := service.Create(ctx, "alice@example.com", "First")
	second, _ := service.Create(ctx, "alice@example.com", "Second")

	if err := service.Delete(ctx, first.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	remaining, err := service.List(ctx, "alice@example.com")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(remaining) != 1 || remaining[0].ID != second.ID {
		t.Fatalf("expected only second todo to remain, got %+v", remaining)
	}

	if err := service.Clear(ctx, "alice@example.com"); err != nil {
		t.Fatalf("clear failed: %v", err)
	}

	remaining, err = service.List(ctx, "alice@example.com")
	if err != nil {
		t.Fatalf("list after clear failed: %v", err)
	}
	if len(remaining) != 0 {
		t.Fatalf("expected 0 todos after clear, got %d", len(remaining))
	}
}

// TestTodoServiceValidateInput covers invalid IDs and payloads.
func TestTodoServiceValidateInput(t *testing.T) {
	ctx := context.Background()
	service := NewTodoService(newMemoryTodoRepo(), fixedNow)

	if _, err := service.Create(ctx, "", ""); err != ErrInvalidTodoInput {
		t.Fatalf("expected ErrInvalidTodoInput for empty create, got %v", err)
	}

	if _, err := service.Update(ctx, "invalid-id", TodoUpdate{Title: strPtr("x")}); err != ErrInvalidTodoID {
		t.Fatalf("expected ErrInvalidTodoID for update, got %v", err)
	}

	if err := service.Delete(ctx, "invalid-id"); err != ErrInvalidTodoID {
		t.Fatalf("expected ErrInvalidTodoID for delete, got %v", err)
	}
}

func strPtr(value string) *string {
	return &value
}
