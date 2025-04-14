package main

import (
	"database/sql"
	"log"
	"os"

	"StartupPCConfigurator/internal/config/handlers"
	"StartupPCConfigurator/internal/config/repository"
	"StartupPCConfigurator/internal/config/usecase"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	// Подключения для БД, миграций и т.д.
)

func main() {
	// 1. Инициализируем соединение с БД
	dbURL := "postgres://user:password@localhost:5432/authdb?sslmode=disable"
	db, err := sql.Open("postgres", dbURL)
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

	// Можно добавить middleware для логирования, CORS, авторизации:
	// r.Use(CORSMiddleware(), AuthMiddleware(), ...)

	// 5. Инициализируем хендлеры (можно одним объектом или несколькими)
	// Здесь передаём сервис (бизнес-логику) в конструктор хендлера
	h := handlers.NewConfigHandler(service)

	// 6. Регистрация роутов
	// Эндпоинты, описанные в openapi.yaml (/config/components и т.д.)
	r.GET("/config/components", h.GetComponents)
	r.POST("/config/newconfig", h.CreateConfig)
	r.GET("/config/userconf", h.GetUserConfigs)
	r.PUT("/config/newconfig/:configId", h.UpdateConfig)
	r.DELETE("/config/newconfig/:configId", h.DeleteConfig)

	// 7. Запуск сервера на порте (например, 8081)
	port := os.Getenv("CONFIG_SERVICE_PORT")
	if port == "" {
		port = "8081"
	}
	r.Run(":" + port)
}

//func initDB() (*sql.DB, error) {
//	// Пример: чтение DSN из env или конфига
//	dsn := os.Getenv("DB_DSN")
//	// Подключение к Postgres (github.com/lib/pq)
//	db, err := sql.Open("postgres", dsn)
//	if err != nil {
//		return nil, err
//	}
//	// Проверка соединения
//	if err := db.Ping(); err != nil {
//		return nil, err
//	}
//	return db, nil
//}
