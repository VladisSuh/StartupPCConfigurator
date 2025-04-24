package main

import (
	"StartupPCConfigurator/internal/auth/handlers"
	"StartupPCConfigurator/internal/auth/repository"
	"StartupPCConfigurator/internal/auth/service"
	"StartupPCConfigurator/pkg/middleware"
	"os"

	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	// 1. Конфигурация (строка подключения, секрет JWT и т.д.)
	jwtSecret := os.Getenv("JWT_SECRET") // Обычно берется из env

	// 2. Подключение к базе данных
	dsn := os.Getenv("DB_CONN_STR")
	if dsn == "" {
		log.Fatal("DB_CONN_STR не задан")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}
	defer db.Close()

	// 3. Инициализация репозитория и сервиса
	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo, jwtSecret)
	handler := handlers.NewHandler(authSvc)

	// 4. Настройка роутера (с использованием Gin)
	router := gin.Default()
	authGroup := router.Group("/")
	{
		authGroup.POST("/register", handler.Register)
		authGroup.POST("/login", handler.Login)
		authGroup.POST("/refresh", handler.Refresh)
		authGroup.POST("/forgot_password", handler.ForgotPassword)
		authGroup.POST("/reset_password", handler.ResetPassword)
		authGroup.POST("/verify_email", handler.VerifyEmail)

		// Защищённые (требуют access_token)
		authGroup.Use(middleware.AuthMiddleware(jwtSecret))
		{
			authGroup.GET("/me", handler.Me)
			authGroup.POST("/logout", handler.Logout)
			authGroup.DELETE("/delete", handler.DeleteAccount)
		}
	}

	// 5. Запуск сервера
	log.Println("Auth service running on :8001")
	if err := router.Run(":8001"); err != nil {
		log.Fatal("Server failed:", err)
	}
}
