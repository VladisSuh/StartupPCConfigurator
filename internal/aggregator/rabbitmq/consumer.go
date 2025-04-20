package rabbitmq

import (
	"context"
	"encoding/json"
	"log"

	"github.com/streadway/amqp"

	"StartupPCConfigurator/internal/aggregator/usecase"
)

// Предположим, у вас в usecase определена структура:
//
//	type ShopUpdateMsg struct { JobID, ShopID int64; Type string }
//
// или что‑то подобное.
// Если она в другом пакете — поправьте импорт.
type ShopUpdateMsg = usecase.ShopUpdateMsg

// StartAggregatorConsumer слушает очередь "shop_update" и вызывает ProcessShopUpdate
func StartAggregatorConsumer(
	ch *amqp.Channel,
	updUC usecase.UpdateUseCase, // <— теперь UpdateUseCase
	logger *log.Logger,
) error {
	q, err := ch.QueueDeclare("shop_update", true, false, false, false, nil)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	logger.Println("Shop‑update consumer running...")
	for d := range msgs {
		var msg usecase.ShopUpdateMsg
		json.Unmarshal(d.Body, &msg)
		updUC.ProcessShopUpdate(context.Background(), msg.JobID, msg.ShopID)
	}
	return nil
}
