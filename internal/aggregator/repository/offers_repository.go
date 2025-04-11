package repository

import (
	"context"
	"database/sql"
	_ "fmt"

	_ "github.com/lib/pq" // драйвер PostgreSQL

	"StartupPCConfigurator/internal/domain"
)

type offersRepository struct {
	db *sql.DB
}

// Интерфейс, на который ссылается usecase
type OffersRepository interface {
	FetchOffers(ctx context.Context, filter domain.OffersFilter) ([]domain.Offer, error)
}

// Конструктор
func NewOffersRepository(connStr string) (OffersRepository, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	// Проверить подключение, опционально
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &offersRepository{db: db}, nil
}

func (r *offersRepository) FetchOffers(ctx context.Context, filter domain.OffersFilter) ([]domain.Offer, error) {
	// Пример простого запроса
	// Предположим, таблица "offers" хранит (component_id, shop_id, price, availability, etc.)
	query := `
        SELECT shop_id, shop_name, price, currency, availability, url
        FROM offers
        WHERE component_id = $1
    `
	args := []interface{}{filter.ComponentID}

	// Примитивная логика сортировки
	if filter.Sort == "priceAsc" {
		query += " ORDER BY price ASC"
	} else if filter.Sort == "priceDesc" {
		query += " ORDER BY price DESC"
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Offer
	for rows.Next() {
		var offer domain.Offer
		err := rows.Scan(
			&offer.ShopID,
			&offer.ShopName,
			&offer.Price,
			&offer.Currency,
			&offer.Availability,
			&offer.URL,
		)
		if err != nil {
			return nil, err
		}
		// componentId тоже можно добавлять, если нужно в ответе
		offer.ComponentID = filter.ComponentID

		result = append(result, offer)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return result, nil
}
