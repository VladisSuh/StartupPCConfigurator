package main

import (
	"StartupPCConfigurator/internal/auth/handlers"
	"StartupPCConfigurator/internal/auth/repository"
	"StartupPCConfigurator/internal/auth/service"
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	// 1. Конфигурация (строка подключения, секрет JWT и т.д.)
	dbURL := "postgres://user:password@localhost:5432/authdb?sslmode=disable"
	jwtSecret := "your_secret_key" // Обычно берется из env

	// 2. Подключение к базе данных
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Database connection error:", err)
	}
	defer db.Close()

	// 3. Инициализация репозитория и сервиса
	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo, jwtSecret)
	handler := handlers.NewHandler(authSvc)

	// 4. Настройка роутера (с использованием Gin)
	router := gin.Default()
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", handler.Register)
		authGroup.POST("/login", handler.Login)
	}

	// 5. Запуск сервера
	log.Println("Auth service running on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Server failed:", err)
	}
}
