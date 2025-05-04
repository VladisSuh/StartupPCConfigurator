package rabbitmq

import (
	"StartupPCConfigurator/internal/aggregator/usecase"
	"context"
	"encoding/json"
	"github.com/streadway/amqp"
	"log"
	"os"
)

type ImportMsg struct {
	FilePath string `json:"filePath"`
}

func StartImportConsumer(ch *amqp.Channel, uc usecase.OffersUseCase, logger *log.Logger) error {
	q, _ := ch.QueueDeclare("price_list_import", true, false, false, false, nil)
	msgs, _ := ch.Consume(q.Name, "", true, false, false, false, nil)

	for d := range msgs {
		var m ImportMsg
		if err := json.Unmarshal(d.Body, &m); err != nil {
			logger.Printf("import: invalid message: %v", err)
			continue
		}
		// Открываем файл по пути:
		f, err := os.Open(m.FilePath)
		if err != nil {
			logger.Printf("import: cannot open %s: %v", m.FilePath, err)
			continue
		}
		if err := uc.ImportPriceList(context.Background(), f); err != nil {
			logger.Printf("import: failed for %s: %v", m.FilePath, err)
		} else {
			logger.Printf("import: success for %s", m.FilePath)
		}
		f.Close()
	}
	return nil
}
