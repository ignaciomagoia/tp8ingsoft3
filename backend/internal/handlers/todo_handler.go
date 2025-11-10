package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ignaciomagoia/tp6ingdesoft/backend/internal/services"
)

// TodoHandler exposes HTTP handlers for todo operations.
type TodoHandler struct {
	todos *services.TodoService
}

// NewTodoHandler builds a new TodoHandler instance.
func NewTodoHandler(todos *services.TodoService) *TodoHandler {
	return &TodoHandler{todos: todos}
}

// ListTodos retrieves todos filtered by email if provided.
func (h *TodoHandler) ListTodos(c *gin.Context) {
	email := c.Query("email")
	todos, err := h.todos.List(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al obtener tareas"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"todos": todos})
}

type createTodoRequest struct {
	Email string `json:"email"`
	Title string `json:"title"`
}

// CreateTodo stores a new todo.
func (h *TodoHandler) CreateTodo(c *gin.Context) {
	var payload createTodoRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos invalidos"})
		return
	}

	todo, err := h.todos.Create(c.Request.Context(), payload.Email, payload.Title)
	switch {
	case err == nil:
		c.JSON(http.StatusCreated, gin.H{"todo": todo})
	case errors.Is(err, services.ErrInvalidTodoInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "email y titulo son requeridos"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al crear tarea"})
	}
}

type updateTodoRequest struct {
	Title     *string `json:"title"`
	Completed *bool   `json:"completed"`
}

// UpdateTodo modifies an existing todo.
func (h *TodoHandler) UpdateTodo(c *gin.Context) {
	id := c.Param("id")

	var payload updateTodoRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos invalidos"})
		return
	}

	todo, err := h.todos.Update(c.Request.Context(), id, services.TodoUpdate{
		Title:     payload.Title,
		Completed: payload.Completed,
	})
	switch {
	case err == nil:
		c.JSON(http.StatusOK, gin.H{"todo": todo})
	case errors.Is(err, services.ErrInvalidTodoInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "nada para actualizar"})
	case errors.Is(err, services.ErrInvalidTodoID):
		c.JSON(http.StatusBadRequest, gin.H{"error": "id invalido"})
	case errors.Is(err, services.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "tarea no encontrada"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al actualizar tarea"})
	}
}

// DeleteTodo removes a todo by ID.
func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	id := c.Param("id")

	err := h.todos.Delete(c.Request.Context(), id)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, gin.H{"message": "tarea eliminada"})
	case errors.Is(err, services.ErrInvalidTodoID):
		c.JSON(http.StatusBadRequest, gin.H{"error": "id invalido"})
	case errors.Is(err, services.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "tarea no encontrada"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al eliminar tarea"})
	}
}

// ClearTodos removes todos optionally filtered by email.
func (h *TodoHandler) ClearTodos(c *gin.Context) {
	email := c.Query("email")
	if err := h.todos.Clear(c.Request.Context(), email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al limpiar tareas"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tareas eliminadas"})
}
