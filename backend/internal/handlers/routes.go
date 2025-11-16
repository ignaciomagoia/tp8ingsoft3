package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RouterConfig allows customising router construction (handy for tests).
type RouterConfig struct {
	Middlewares []gin.HandlerFunc
}

// SetupRouter wires handlers with the HTTP routes.
func SetupRouter(auth *AuthHandler, todos *TodoHandler, cfg RouterConfig) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	if len(cfg.Middlewares) > 0 {
		router.Use(cfg.Middlewares...)
	}

	// Responder preflight OPTIONS para cualquier ruta para que el middleware CORS
	// pueda contestar la solicitud antes de que se bloquee por falta de handler.
	router.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.POST("/register", auth.Register)
	router.POST("/login", auth.Login)
	router.GET("/users", auth.ListUsers)
	router.DELETE("/users", auth.ClearUsers)

	router.GET("/todos", todos.ListTodos)
	router.POST("/todos", todos.CreateTodo)
	router.PUT("/todos/:id", todos.UpdateTodo)
	router.DELETE("/todos/:id", todos.DeleteTodo)
	router.DELETE("/todos", todos.ClearTodos)

	return router
}
