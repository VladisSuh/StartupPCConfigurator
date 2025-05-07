package main

import (
	"StartupPCConfigurator/internal/aggregator/rabbitmq"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"

	"github.com/gin-gonic/gin"

	"StartupPCConfigurator/internal/aggregator/handlers"
	"StartupPCConfigurator/internal/aggregator/parser/dns"
	"StartupPCConfigurator/internal/aggregator/repository"
	"StartupPCConfigurator/internal/aggregator/usecase"
	// например, "_ github.com/lib/pq" если нужно драйвер для PostgreSQL
)

func main() {
	logger := log.Default()

	// === 1. Подключение к БД (как раньше) ===
	dbConnStr := os.Getenv("DB_CONN_STR")
	for i := 0; i < 3; i++ {
		if dbConnStr == "" {
			dbConnStr = "postgres://postgres:postgres@postgres:5432/postgres_db?sslmode=disable"
		}
		log.Println("Postgres подключается, ожидайте ещё 5 секунд...")
		time.Sleep(5 * time.Second)
	}
	repo, err := repository.NewRepository(dbConnStr)
	if err != nil {
		logger.Fatalf("Error creating repository: %v", err)
	}

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

	const exchange = "aggregator-ex"
	if err := ch.ExchangeDeclare(exchange, "direct", true, false, false, false, nil); err != nil {
		logger.Fatalf("ExchangeDeclare error: %v", err)
	}
	q, err := ch.QueueDeclare("price.changed", true, false, false, false, nil)
	if err != nil {
		logger.Fatalf("QueueDeclare error: %v", err)
	}
	if err := ch.QueueBind(q.Name, "price.changed", exchange, false, nil); err != nil {
		logger.Fatalf("QueueBind error: %v", err)
	}

	publisher := rabbitmq.NewAggregatorPublisher(ch, logger, exchange)

	// === 2. Инициализация HTTP‑UseCase и Handler ===
	offersUC := usecase.NewOffersUseCase(repo, publisher, logger)
	offersHandler := handlers.NewOffersHandler(offersUC)

	// === 4. Инициализация Parser, Publisher и Update‑UseCase ===
	// DNS‑парсер
	dnsParser := dns.NewDNSParser(logger)
	// Update‑UseCase, который обрабатывает очередь shop_update
	updateUC := usecase.NewUpdateUseCase(repo, publisher, dnsParser, logger)

	// === 5. Старт Consumer’а в фоне ===
	go func() {
		if err := rabbitmq.StartAggregatorConsumer(ch, updateUC, logger); err != nil {
			logger.Fatalf("Consumer error: %v", err)
		}
	}()

	go func() {
		if err := rabbitmq.StartImportConsumer(ch, offersUC, logger); err != nil {
			logger.Fatalf("Import consumer error: %v", err)
		}
	}()

	// === 7. Запуск HTTP‑сервера ===
	r := gin.Default()
	r.GET("/offers", offersHandler.GetOffers)
	r.POST("/offers/import", offersHandler.UploadPriceList)

	port := os.Getenv("AGGREGATOR_PORT")
	if port == "" {
		port = "8003"
	}
	logger.Printf("Aggregator service running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		logger.Fatalf("Error starting server: %v", err)
	}
}
