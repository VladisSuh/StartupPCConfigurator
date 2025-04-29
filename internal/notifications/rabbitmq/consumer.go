package rabbitmq

import (
	aggUc "StartupPCConfigurator/internal/aggregator/usecase"
	notifUC "StartupPCConfigurator/internal/notifications/usecase"
	"context"
	"encoding/json"
	"github.com/streadway/amqp"
	"log"
)

func StartNotificationsConsumer(ch *amqp.Channel, uc notifUC.NotificationUseCase, logger *log.Logger) error {
	q, _ := ch.QueueDeclare("price.changed", true, false, false, false, nil)
	msgs, _ := ch.Consume(q.Name, "", true, false, false, false, nil)
	for d := range msgs {
		var msg aggUc.PriceChangedMsg
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			logger.Println("Bad message:", err)
			continue
		}
		// on each price.change, уведомляем всех подписавшихся пользователей
		uc.HandlePriceChange(context.Background(), msg)
	}
	return nil
}
