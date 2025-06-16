package usecase

import (
	"StartupPCConfigurator/internal/domain"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/xuri/excelize/v2"
)

// OffersRepository — только для GET /offers
type OffersRepository interface {
	FetchOffers(ctx context.Context, filter domain.OffersFilter) ([]domain.Offer, error)

	GetOfferPrice(ctx context.Context, componentID string, shopID int64) (float64, error)
	GetMinPrice(ctx context.Context, componentID string) (float64, string, error)
	UpsertOffer(ctx context.Context, componentID string, shopID int64, price float64, availability, url string) error
	InsertPriceHistory(ctx context.Context, componentID string, shopID int64, price float64) error
	GetShopIDByCode(ctx context.Context, code string) (int64, error)
}

type OffersUseCase interface {
	GetOffers(ctx context.Context, filter domain.OffersFilter) ([]domain.Offer, error)
	GetMinPrice(ctx context.Context, componentID string) (float64, string, error)
	ImportPriceList(ctx context.Context, records io.Reader) error
}

type offersUseCase struct {
	repo      OffersRepository
	publisher Publisher
	logger    *log.Logger
}

func NewOffersUseCase(
	repo OffersRepository,
	pub Publisher,
	logger *log.Logger,
) OffersUseCase {
	return &offersUseCase{
		repo:      repo,
		publisher: pub,
		logger:    logger,
	}
}

func (uc *offersUseCase) GetOffers(ctx context.Context, filter domain.OffersFilter) ([]domain.Offer, error) {
	if filter.ComponentID == "" {
		return nil, errors.New("componentId is required")
	}
	return uc.repo.FetchOffers(ctx, filter)
}

func (uc *offersUseCase) GetMinPrice(ctx context.Context, componentID string) (float64, string, error) {
	return uc.repo.GetMinPrice(ctx, componentID)
}

func (uc *offersUseCase) ImportPriceList(ctx context.Context, r io.Reader) error {
	// 1. Читаем Excel
	f, err := excelize.OpenReader(r)
	if err != nil {
		return fmt.Errorf("cannot read excel: %w", err)
	}
	rows, err := f.GetRows("Sheet1") // или по имени/индексу
	if err != nil {
		return fmt.Errorf("cannot get rows: %w", err)
	}

	// 2. Собираем записи
	//var recs []ImportRecord

	for i, row := range rows {
		if i == 0 {
			continue
		}

		code := row[1] // из XLSX: ячейка с кодом магазина, например "DNS"
		// --- вместо ParseInt: получаем реальный shop_id из БД ---
		shopID, err := uc.repo.GetShopIDByCode(ctx, code)
		if err != nil {
			uc.logger.Printf("unknown shop code %q: %v", code, err)
			continue
		}

		price, _ := strconv.ParseFloat(row[2], 64)

		// 1) получить старую цену
		old, _ := uc.repo.GetOfferPrice(ctx, row[0], shopID)
		// 2) upsert + history
		_ = uc.repo.UpsertOffer(ctx, row[0], shopID, price, row[4], row[5])
		_ = uc.repo.InsertPriceHistory(ctx, row[0], shopID, price)

		// 3) если изменилась — публикуем в RabbitMQ
		if price != old {
			if err := uc.publisher.PublishPriceChanged(row[0], shopID, old, price); err != nil {
				uc.logger.Printf("publish failed: %v", err)
			}
		}
	}
	return nil
}
