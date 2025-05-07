package main

import (
	"context"
	"log"
	"os"
	"time"

	"database/sql"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/streadway/amqp"

	"StartupPCConfigurator/internal/notifications/handlers"
	"StartupPCConfigurator/internal/notifications/rabbitmq"
	"StartupPCConfigurator/internal/notifications/repository"
	"StartupPCConfigurator/internal/notifications/usecase"
	"StartupPCConfigurator/pkg/middleware"
)

func main() {
	// Logger
	logger := log.Default()

	// === 1. Configuration from env ===
	dbConnStr := os.Getenv("DB_CONN_STR")
	for i := 0; i < 3; i++ {
		if dbConnStr == "" {
			dbConnStr = "postgres://postgres:postgres@localhost:5432/postgres_db?sslmode=disable"
		}
		log.Println("Postgres подключается, ожидайте ещё 5 секунд...")
		time.Sleep(5 * time.Second)
	}

	redisAddr := os.Getenv("REDIS_URL")
	for i := 0; i < 3; i++ {
		if redisAddr == "" {
			redisAddr = "localhost:6379"
		}
		log.Println("Redis подключается, ожидайте ещё 5 секунд...")
		time.Sleep(5 * time.Second)
	}

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "secret_key"
	}

	httpPort := os.Getenv("NOTIFICATIONS_PORT")
	if httpPort == "" {
		httpPort = "8004"
	}

	// === 2. Connect to Postgres ===
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		logger.Fatalf("Failed to open Postgres: %v", err)
	}
	if err := db.Ping(); err != nil {
		logger.Fatalf("Failed to ping Postgres: %v", err)
	}

	// === 3. Connect to Redis ===
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // no password set
		DB:       0,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Fatalf("Failed to ping Redis: %v", err)
	}

	// === 4. Connect to RabbitMQ ===
	var conn *amqp.Connection
	for i := 0; i < 5; i++ {
		conn, err = amqp.Dial(rabbitURL)
		if err == nil {
			break
		}
		logger.Println("RabbitMQ connecting, retrying in 2s...")
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		logger.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		logger.Fatalf("Failed to open RabbitMQ channel: %v", err)
	}
	defer ch.Close()

	// === 5. Init Repository, Cache, UseCase, Handler ===
	notifRepo := repository.NewNotificationRepository(db)
	notifCache := repository.NewNotificationCache(rdb)
	notifUC := usecase.NewNotificationUseCase(notifRepo, notifCache, logger)
	notifHandler := handlers.NewHandler(notifUC)

	// === 6. Start RabbitMQ consumer ===
	go func() {
		if err := rabbitmq.StartNotificationsConsumer(ch, notifUC, logger); err != nil {
			logger.Fatalf("Notifications consumer error: %v", err)
		}
	}()

	// === 7. HTTP Server (Gin) ===
	r := gin.Default()
	r.Use(cors.Default())

	// Protected routes
	n := r.Group("/notifications")
	n.Use(middleware.AuthMiddleware(jwtSecret))
	{
		n.GET("/count", notifHandler.UnreadCount)
		n.GET("", notifHandler.List)
		n.POST("/:id/read", notifHandler.MarkRead)
	}

	subs := r.Group("/subscriptions")
	subs.Use(middleware.AuthMiddleware(jwtSecret))
	{
		subs.POST("", notifHandler.Subscribe)
		subs.DELETE("/:componentId", notifHandler.Unsubscribe)
	}

	logger.Printf("Notifications service listening on :%s", httpPort)
	if err := r.Run(":" + httpPort); err != nil {
		logger.Fatalf("Failed to run HTTP server: %v", err)
	}
}
