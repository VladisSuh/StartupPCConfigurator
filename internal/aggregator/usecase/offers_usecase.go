package usecase

import (
	"StartupPCConfigurator/internal/domain"
	"context"
	"errors"
	"fmt"
	"github.com/xuri/excelize/v2"
	"io"
	"strconv"
)

// OffersRepository — только для GET /offers
type OffersRepository interface {
	FetchOffers(ctx context.Context, filter domain.OffersFilter) ([]domain.Offer, error)
	BulkUpsertOffers(ctx context.Context, recs []ImportRecord) error
}

type OffersUseCase interface {
	GetOffers(ctx context.Context, filter domain.OffersFilter) ([]domain.Offer, error)
	ImportPriceList(ctx context.Context, records io.Reader) error
}

type offersUseCase struct {
	repo OffersRepository
}

func NewOffersUseCase(r OffersRepository) OffersUseCase {
	return &offersUseCase{repo: r}
}

func (uc *offersUseCase) GetOffers(ctx context.Context, filter domain.OffersFilter) ([]domain.Offer, error) {
	if filter.ComponentID == "" {
		return nil, errors.New("componentId is required")
	}
	return uc.repo.FetchOffers(ctx, filter)
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
	var recs []ImportRecord

	for i, row := range rows {
		if i == 0 {
			continue
		} // пропускаем header
		price, _ := strconv.ParseFloat(row[2], 64) // колонка C
		recs = append(recs, ImportRecord{
			ComponentID:  row[0], // A
			ShopCode:     row[1], // B
			Price:        price,
			Currency:     row[3], // D
			Availability: row[4], // E
			URL:          row[5], // F
		})
	}

	// 3. Для каждого делаем bulk-upsert
	return uc.repo.BulkUpsertOffers(ctx, recs)
}
