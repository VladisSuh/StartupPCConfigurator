package main

import (
	"StartupPCConfigurator/internal/aggregator/rabbitmq"
	"github.com/streadway/amqp"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"StartupPCConfigurator/internal/aggregator/handlers"
	"StartupPCConfigurator/internal/aggregator/repository"
	"StartupPCConfigurator/internal/aggregator/usecase"
	// например, "_ github.com/lib/pq" если нужно драйвер для PostgreSQL
)

func main() {
	logger := log.Default()

	// === 1. Подключение к БД (как раньше) ===
	dbConnStr := os.Getenv("DB_CONN_STR")
	if dbConnStr == "" {
		dbConnStr = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}
	repo, err := repository.NewOffersRepository(dbConnStr)
	if err != nil {
		logger.Fatalf("Error creating repository: %v", err)
	}

	// === 2. Инициализация UseCase ===
	offersUC := usecase.NewOffersUseCase(repo, logger)

	// === 3. Подключение к RabbitMQ ===
	var conn *amqp.Connection
	for i := 0; i < 10; i++ {
		rabbitURL := os.Getenv("RABBITMQ_URL")
		if rabbitURL == "" {
			rabbitURL = "amqp://guest:guest@localhost:5672/"
		}
		conn, err = amqp.Dial(rabbitURL)
		if err == nil {
			log.Println("Успешно подключились к RabbitMQ")
			break
		}
		log.Println("RabbitMQ подключается, ожидайте ещё 5 секунд...")
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		logger.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Создаём канал
	ch, err := conn.Channel()
	if err != nil {
		logger.Fatalf("failed to open channel: %v", err)
	}
	defer ch.Close()

	// === 4. Инициализация Consumer’а (подписчика) ===
	go func() {
		// например, объявим очередь aggregator_update, которую будем слушать
		err := rabbitmq.StartAggregatorConsumer(ch, offersUC, logger)
		if err != nil {
			logger.Fatalf("consumer error: %v", err)
		}
	}()

	// === 5. Инициализация HTTP-сервера ===
	r := gin.Default()

	offersHandler := handlers.NewOffersHandler(offersUC)
	r.GET("/offers", offersHandler.GetOffers)

	port := os.Getenv("AGGREGATOR_PORT")
	if port == "" {
		port = "8082"
	}
	logger.Printf("Aggregator service running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		logger.Fatalf("Error starting server: %v", err)
	}
}
