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

type Publisher interface {
	PublishPriceChanged(componentID string, shopID int64, price float64) error
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

func (p *AggregatorPublisher) PublishPriceChanged(
	componentID string,
	shopID int64,
	oldPrice, newPrice float64,
) error {
	// Собираем структуру, совпадающую с JSON-схемой
	evt := struct {
		ComponentID string  `json:"componentId"`
		ShopID      int64   `json:"shopId"`
		OldPrice    float64 `json:"oldPrice"`
		NewPrice    float64 `json:"newPrice"`
	}{
		ComponentID: componentID,
		ShopID:      shopID,
		OldPrice:    oldPrice,
		NewPrice:    newPrice,
	}
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	// Публикуем в ту же очередь / exchange, где слушает Notifications Service
	return p.ch.Publish(
		p.exName,        // exchange
		"price.changed", // routing key
		false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
