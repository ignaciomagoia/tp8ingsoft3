package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/ignaciomagoia/tp6ingdesoft/backend/internal/handlers"
	"github.com/ignaciomagoia/tp6ingdesoft/backend/internal/services"
)

func getAllowedOrigins() []string {
	env := os.Getenv("FRONT_ORIGINS")
	log.Printf("FRONT_ORIGINS env raw: %q", env)
	if env == "" {
		return []string{"http://localhost:5173"}
	}
	parts := strings.Split(env, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
			log.Printf("Allowed origin loaded: %s", s)
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
	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	db := client.Database(dbName)

	userRepo := services.NewMongoUserRepository(db.Collection("users"))
	todoRepo := services.NewMongoTodoRepository(db.Collection("todos"))

	userService := services.NewUserService(userRepo)
	todoService := services.NewTodoService(todoRepo, time.Now)

	authHandler := handlers.NewAuthHandler(userService)
	todoHandler := handlers.NewTodoHandler(todoService)

	router := handlers.SetupRouter(authHandler, todoHandler, handlers.RouterConfig{})

	corsCfg := cors.Config{
		AllowOrigins:     getAllowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	log.Printf("CORS final config origins: %v", corsCfg.AllowOrigins)
	router.Use(cors.New(corsCfg))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("no se pudo iniciar el servidor: %v", err)
	}
}
