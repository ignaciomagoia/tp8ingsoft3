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
	log.Printf("[CORS] FRONT_ORIGINS desde env: %q", env)
	if env == "" {
		log.Printf("[CORS] ⚠ FRONT_ORIGINS vacío, usando default: http://localhost:5173")
		return []string{"http://localhost:5173"}
	}
	parts := strings.Split(env, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
			log.Printf("[CORS] ✅ Origen permitido: %q", s)
		}
	}
	log.Printf("[CORS] Total de orígenes permitidos: %d", len(out))
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

	// Configurar CORS ANTES de crear el router
	allowedOrigins := getAllowedOrigins()

	// Crear un mapa para búsqueda rápida de orígenes permitidos
	allowedOriginsMap := make(map[string]bool)
	for _, origin := range allowedOrigins {
		allowedOriginsMap[origin] = true
	}

	corsCfg := cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// Si no hay origen (petición del mismo origen), permitir
			if origin == "" {
				return true
			}
			// Permitir si el origen está en la lista
			if allowedOriginsMap[origin] {
				log.Printf("[CORS] ✅ Origen permitido: %q", origin)
				return true
			}
			log.Printf("[CORS] ❌ Origen bloqueado: %q (permitidos: %v)", origin, allowedOrigins)
			return false
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	log.Printf("[CORS] Configuración aplicada con %d orígenes permitidos", len(allowedOrigins))

	router := handlers.SetupRouter(authHandler, todoHandler, handlers.RouterConfig{})

	// Aplicar CORS como PRIMER middleware (crítico para que funcione correctamente)
	router.Use(cors.New(corsCfg))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("[SERVER] Iniciando servidor en puerto %s", port)
	log.Printf("[SERVER] CORS configurado para orígenes: %v", allowedOrigins)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("no se pudo iniciar el servidor: %v", err)
	}
}
