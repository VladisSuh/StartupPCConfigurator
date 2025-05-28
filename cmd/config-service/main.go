package main

import (
	"database/sql"
	"log"
	"os"

	"StartupPCConfigurator/internal/config/handlers"
	"StartupPCConfigurator/internal/config/repository"
	"StartupPCConfigurator/internal/config/usecase"
	"StartupPCConfigurator/pkg/middleware"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	// Подключения для БД, миграций и т.д.
)

func main() {
	// 1. Инициализируем соединение с БД
	dsn := os.Getenv("DB_CONN_STR")
	if dsn == "" {
		log.Fatal("DB_CONN_STR не задан")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}
	defer db.Close()

	// 2. Создаём репозиторий
	repo := repository.NewConfigRepository(db)

	// 3. Создаём уровень бизнес-логики (usecase)
	service := usecase.NewConfigService(repo)

	// 4. Создаём Gin-роутер
	r := gin.Default()

	// 5. Подключаем middleware
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET не задан")
	}

	auth := middleware.AuthMiddleware(jwtSecret)
	h := handlers.NewConfigHandler(service)

	// 6. Публичные ручки
	r.GET("/components", h.GetComponents)
	r.POST("/compatible", h.GetCompatibleComponentsMulti)
	r.GET("/usecases", h.ListUseCases)
	r.GET("/usecase/:name", h.GetUseCaseBuild)
	r.GET("/brands", h.GetBrands)

	// 7. Защищённые ручки
	api := r.Group("/", auth)
	{
		api.POST("/newconfig", h.CreateConfig)
		api.GET("/userconf", h.GetUserConfigs)
		api.PUT("/newconfig/:configId", h.UpdateConfig)
		api.DELETE("/newconfig/:configId", h.DeleteConfig)
	}

	// 7. Запуск сервера на порте (например, 8081)
	port := os.Getenv("CONFIG_SERVICE_PORT")
	if port == "" {
		port = "8002"
	}
	r.Run(":" + port)
}
