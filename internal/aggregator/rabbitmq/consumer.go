package rabbitmq

import (
	"context"
	"encoding/json"
	"log"

	"github.com/streadway/amqp"

	"StartupPCConfigurator/internal/aggregator/usecase"
	"StartupPCConfigurator/internal/domain"
)

// StartAggregatorConsumer — пример подписчика, который слушает очередь "aggregator_update"
// и при получении сообщения запускает какую-то логику (например, обновить каталог).
func StartAggregatorConsumer(ch *amqp.Channel, offersUC usecase.OffersUseCase, logger *log.Logger) error {
	// 1. Объявляем очередь (если её нет, RabbitMQ создаст)
	q, err := ch.QueueDeclare(
		"aggregator_update", // имя очереди
		true,                // durable
		false,               // autoDelete
		false,               // exclusive
		false,               // noWait
		nil,                 // аргументы
	)
	if err != nil {
		return err
	}

	// 2. Подписываемся на очередь
	msgs, err := ch.Consume(
		q.Name,
		"",    // consumer
		true,  // autoAck (для упрощения)
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	// 3. Запускаем цикл чтения сообщений
	logger.Println("Aggregator consumer is running, waiting for messages...")
	for d := range msgs {
		// d.Body — []byte
		logger.Printf("Received message: %s", string(d.Body))

		// Допустим, у нас JSON-структура:
		var event domain.UpdateEvent
		if err := json.Unmarshal(d.Body, &event); err != nil {
			logger.Printf("Error unmarshalling message: %v", err)
			continue
		}

		// Вызываем бизнес-логику в usecase — напр., запустить обновление
		// Здесь можно передать context с таймаутом
		ctx := context.Background()
		if err := offersUC.ProcessUpdateEvent(ctx, event); err != nil {
			logger.Printf("Error processing update event: %v", err)
		}
	}
	return nil
}
