package usecase

// Publisher — абстракция для публикации событий в RabbitMQ
type Publisher interface {
	// PublishPriceChanged генерирует событие price.changed
	PublishPriceChanged(componentID string, shopID int64, oldPrice, newPrice float64) error
}
