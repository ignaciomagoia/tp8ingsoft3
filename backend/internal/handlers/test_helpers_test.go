package handlers

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/ignaciomagoia/tp6ingdesoft/backend/internal/services"
)

type memoryUserRepo struct {
	mu    sync.Mutex
	users map[string]services.User
}

func newMemoryUserRepo() *memoryUserRepo {
	return &memoryUserRepo{users: make(map[string]services.User)}
}

func (m *memoryUserRepo) FindByEmail(_ context.Context, email string) (services.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, ok := m.users[email]
	if !ok {
		return services.User{}, services.ErrNotFound
	}
	return user, nil
}

func (m *memoryUserRepo) Insert(_ context.Context, user services.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.users[user.Email] = user
	return nil
}

func (m *memoryUserRepo) List(_ context.Context) ([]services.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	users := make([]services.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].Email < users[j].Email
	})
	return users, nil
}

func (m *memoryUserRepo) Clear(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.users = make(map[string]services.User)
	return nil
}

type memoryTodoRepo struct {
	mu    sync.Mutex
	todos map[primitive.ObjectID]services.Todo
}

func newMemoryTodoRepo() *memoryTodoRepo {
	return &memoryTodoRepo{todos: make(map[primitive.ObjectID]services.Todo)}
}

func (m *memoryTodoRepo) List(_ context.Context, email string) ([]services.Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	todos := make([]services.Todo, 0, len(m.todos))
	for _, todo := range m.todos {
		if email == "" || todo.Email == email {
			todos = append(todos, todo)
		}
	}

	sort.Slice(todos, func(i, j int) bool {
		return todos[i].CreatedAt.Before(todos[j].CreatedAt)
	})
	return todos, nil
}

func (m *memoryTodoRepo) Create(_ context.Context, todo services.Todo) (services.Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	todo.ID = primitive.NewObjectID()
	m.todos[todo.ID] = todo
	return todo, nil
}

func (m *memoryTodoRepo) Update(_ context.Context, id primitive.ObjectID, update services.TodoUpdate) (services.Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	todo, ok := m.todos[id]
	if !ok {
		return services.Todo{}, services.ErrNotFound
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
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.todos[id]; !ok {
		return services.ErrNotFound
	}
	delete(m.todos, id)
	return nil
}

func (m *memoryTodoRepo) Clear(_ context.Context, email string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, todo := range m.todos {
		if email == "" || todo.Email == email {
			delete(m.todos, id)
		}
	}
	return nil
}

type testApp struct {
	router *gin.Engine
	users  *memoryUserRepo
	todos  *memoryTodoRepo
}

func newTestApp() *testApp {
	gin.SetMode(gin.TestMode)

	users := newMemoryUserRepo()
	todos := newMemoryTodoRepo()

	userService := services.NewUserService(users)
	todoService := services.NewTodoService(todos, func() time.Time { return fixedTime })

	authHandler := NewAuthHandler(userService)
	todoHandler := NewTodoHandler(todoService)

	router := SetupRouter(authHandler, todoHandler, RouterConfig{})

	return &testApp{
		router: router,
		users:  users,
		todos:  todos,
	}
}

var fixedTime = time.Date(2025, time.January, 1, 10, 0, 0, 0, time.UTC)
