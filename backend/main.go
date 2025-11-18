package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ignaciomagoia/tp6ingdesoft/backend/internal/handlers"
	"github.com/ignaciomagoia/tp6ingdesoft/backend/internal/services"
)

func getAllowedOrigins() []string {
	defaultOrigins := []string{
		"http://localhost:3000",
		"http://localhost:5173",
		"http://127.0.0.1:3000",
		"http://127.0.0.1:5173",
	}

	env := os.Getenv("FRONT_ORIGINS")
	log.Printf("[CORS] FRONT_ORIGINS desde env: %q", env)

	out := make([]string, 0, len(defaultOrigins))
	seen := make(map[string]struct{})

	addOrigin := func(origin string) {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			return
		}
		if _, ok := seen[trimmed]; ok {
			return
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
		log.Printf("[CORS] Origen permitido agregado: %q", trimmed)
	}

	for _, origin := range defaultOrigins {
		addOrigin(origin)
	}

	if env != "" {
		for _, o := range strings.Split(env, ",") {
			addOrigin(o)
		}
	}

	return out
}

func main() {
	ctx := context.Background()

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("MONGO_DB")
	if dbName == "" {
		dbName = services.DefaultDatabaseName
	}

	client, err := services.ConnectMongo(ctx, mongoURI)
	if err != nil {
		log.Fatalf("no se pudo conectar a MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database(dbName)

	userRepo := services.NewMongoUserRepository(db.Collection("users"))
	todoRepo := services.NewMongoTodoRepository(db.Collection("todos"))

	userService := services.NewUserService(userRepo)
	todoService := services.NewTodoService(todoRepo, time.Now)

	authHandler := handlers.NewAuthHandler(userService)
	todoHandler := handlers.NewTodoHandler(todoService)

	allowedOrigins := getAllowedOrigins()

	corsCfg := cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	router := gin.Default()

	// 💥 APLICAR CORS ANTES DEL ROUTER
	router.Use(cors.New(corsCfg))

	// Endpoint de salud
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// APIs
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)

	router.GET("/todos", todoHandler.ListTodos)
	router.POST("/todos", todoHandler.CreateTodo)
	router.PUT("/todos/:id", todoHandler.UpdateTodo)
	router.DELETE("/todos/:id", todoHandler.DeleteTodo)
	router.DELETE("/todos", todoHandler.ClearTodos)

	// Importante: Render usa PORT
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("[SERVER] Corriendo en puerto %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
