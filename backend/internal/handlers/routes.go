package handlers

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// RouterConfig allows customising router construction (handy for tests).
type RouterConfig struct {
	AllowedOrigins []string
}

// SetupRouter wires handlers with the HTTP routes.
func SetupRouter(auth *AuthHandler, todos *TodoHandler, cfg RouterConfig) *gin.Engine {
	router := gin.Default()

	origins := cfg.AllowedOrigins
	if len(origins) == 0 {
		origins = []string{"http://localhost:3000", "http://localhost:3001"}
	}

	corsCfg := cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}
	router.Use(cors.New(corsCfg))

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
