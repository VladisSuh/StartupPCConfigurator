package usecase

import (
	"context"
	"log"
	"strconv"
)

type UpdateUseCase interface {
	ProcessShopUpdate(ctx context.Context, jobID, shopID int64) error
}

type updateUseCase struct {
	repo      Repository
	publisher Publisher
	parser    Parser
	logger    *log.Logger
}

func NewUpdateUseCase(
	repo Repository,
	pub Publisher,
	parser Parser,
	logger *log.Logger,
) UpdateUseCase {
	return &updateUseCase{repo, pub, parser, logger}
}

func (uc *updateUseCase) ProcessShopUpdate(ctx context.Context, jobID, shopID int64) error {
	// 1) Пометить job как running
	if err := uc.repo.UpdateJobStatus(ctx, jobID, "running", nil); err != nil {
		return err
	}

	// 2) Список страниц для парсинга
	items, err := uc.repo.ListShopComponents(ctx, shopID)
	if err != nil {
		uc.repo.UpdateJobStatus(ctx, jobID, "failed", err.Error())
		return err
	}

	// 3) Парсим и пишем в offers + history
	for _, it := range items {
		parsed, err := uc.parser.Parse(ctx, it.URL)
		if err != nil {
			uc.logger.Printf("parser error for %s: %v", it.URL, err)
			continue
		}
		// конвертация цены
		price, err := strconv.ParseFloat(parsed.Price, 64)
		if err != nil {
			uc.logger.Printf("price parse error for %s: %v", parsed.Price, err)
			continue
		}

		// upsert в offers
		if err := uc.repo.UpsertOffer(ctx, it.ComponentID, shopID, price, parsed.Availability, parsed.URL); err != nil {
			uc.logger.Printf("db error upsert offer: %v", err)
			continue
		}
		// запись в историю
		if err := uc.repo.InsertPriceHistory(ctx, it.ComponentID, shopID, price); err != nil {
			uc.logger.Printf("db error insert history: %v", err)
		}

		// опубликовать событие
		if err := uc.publisher.PublishPriceChanged(it.ComponentID, shopID, price); err != nil {
			uc.logger.Printf("publish error: %v", err)
		}
	}

	// 4) Завершить job
	return uc.repo.UpdateJobStatus(ctx, jobID, "done", nil)
}
