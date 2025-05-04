// cmd/importer/main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/streadway/amqp"
)

const (
	defaultRabbitURL = "amqp://guest:guest@localhost:5672/"
	queueName        = "price_list_import"
	processedDirName = "processed"
)

type ImportMsg struct {
	FilePath string `json:"filePath"`
	// Optional: add ShopCode or other metadata if needed
}

func main() {
	// Параметры командной строки
	dir := flag.String("dir", "", "directory with Excel files to import")
	envRabbit := os.Getenv("RABBITMQ_URL")
	if envRabbit == "" {
		envRabbit = defaultRabbitURL
	}
	rabbitURL := flag.String("rabbit", envRabbit, "RabbitMQ connection URL")
	flag.Parse()

	if *dir == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Настраиваем логирование
	logger := log.New(os.Stdout, "importer: ", log.LstdFlags|log.Lshortfile)

	// Подключение к RabbitMQ
	conn, ch := connectRabbit(*rabbitURL, logger)
	defer conn.Close()
	defer ch.Close()

	// Объявляем очередь (если не существует)
	declareQueue(ch, queueName, logger)

	// Создаем папку processed
	processedDir := filepath.Join(*dir, processedDirName)
	if err := os.MkdirAll(processedDir, 0o755); err != nil {
		logger.Fatalf("failed to create processed dir: %v", err)
	}

	// Сканируем директорию и публикуем сообщения
	if err := publishAll(ch, *dir, processedDir, logger); err != nil {
		logger.Fatalf("error publishing messages: %v", err)
	}

	logger.Println("Done publishing all files.")
}

// connectRabbit устанавливает соединение и канал к RabbitMQ
func connectRabbit(rabbitURL string, logger *log.Logger) (*amqp.Connection, *amqp.Channel) {
	var conn *amqp.Connection
	var err error
	for i := 0; i < 5; i++ {
		conn, err = amqp.Dial(rabbitURL)
		if err == nil {
			logger.Println("Connected to RabbitMQ")
			break
		}
		logger.Printf("RabbitMQ connection failed, retrying in 3s: %v", err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		logger.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		logger.Fatalf("failed to open channel: %v", err)
	}
	return conn, ch
}

// declareQueue декларация очереди для публикации
func declareQueue(ch *amqp.Channel, name string, logger *log.Logger) {
	_, err := ch.QueueDeclare(
		name,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		logger.Fatalf("queue declare error: %v", err)
	}
}

// publishAll читает файлы .xlsx из srcDir, публикует сообщения и переносит файлы в dstDir
func publishAll(ch *amqp.Channel, srcDir, dstDir string, logger *log.Logger) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("read dir %s: %w", srcDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".xlsx" {
			continue
		}
		srcPath := filepath.Join(srcDir, entry.Name())
		// Перемещаем файл в processed
		dstPath := filepath.Join(dstDir, entry.Name())

		if err := os.Rename(srcPath, dstPath); err != nil {
			logger.Printf("failed to move %s to processed: %v", entry.Name(), err)
			continue
		}

		msg := ImportMsg{FilePath: dstPath}
		body, err := json.Marshal(msg)
		if err != nil {
			logger.Printf("json marshal failed for %s: %v", srcPath, err)
			continue
		}

		if err := ch.Publish(
			"",        // exchange
			queueName, // routing key
			false,     // mandatory
			false,     // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			},
		); err != nil {
			logger.Printf("publish failed for %s: %v", dstPath, err)
			continue
		}
		logger.Printf("Published import request for %s", entry.Name())
	}
	return nil
}
