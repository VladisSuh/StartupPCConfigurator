// internal/aggregator/rabbitmq/publisher.go
package rabbitmq

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

// AggregatorPublisher — пример структуры, которая умеет отправлять сообщения
type AggregatorPublisher struct {
	ch     *amqp.Channel
	logger *log.Logger
	exName string // имя обменника, если используете Exchange
}

// NewAggregatorPublisher — конструктор
func NewAggregatorPublisher(ch *amqp.Channel, logger *log.Logger, exName string) *AggregatorPublisher {
	return &AggregatorPublisher{
		ch:     ch,
		logger: logger,
		exName: exName,
	}
}

// PublishPriceUpdated — отправить событие "price_updated" в систему
func (p *AggregatorPublisher) PublishPriceUpdated(shopID string, componentID string) error {
	event := map[string]string{
		"event":       "price_updated",
		"shopId":      shopID,
		"componentId": componentID,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = p.ch.Publish(
		p.exName,     // exchange
		"aggregator", // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return err
	}
	p.logger.Printf("Sent price_updated event for shop=%s component=%s", shopID, componentID)
	return nil
}
