package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/ignaciomagoia/tp6ingdesoft/backend/internal/handlers"
	"github.com/ignaciomagoia/tp6ingdesoft/backend/internal/services"
)

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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("no se pudo iniciar el servidor: %v", err)
	}
}
