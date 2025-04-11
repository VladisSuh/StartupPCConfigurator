package usecase

import (
	"context"
	"errors"
	"log"

	"StartupPCConfigurator/internal/domain"
)

type OffersRepository interface {
	FetchOffers(ctx context.Context, filter domain.OffersFilter) ([]domain.Offer, error)
	// при необходимости другие методы (SaveOffer, FetchShops и т.д.)
}

type OffersUseCase interface {
	GetOffers(ctx context.Context, filter domain.OffersFilter) ([]domain.Offer, error)
	ProcessUpdateEvent(ctx context.Context, evt domain.UpdateEvent) error
}

type offersUseCase struct {
	repo   OffersRepository
	logger *log.Logger
}

// Конструктор
func NewOffersUseCase(r OffersRepository, logger *log.Logger) OffersUseCase {
	return &offersUseCase{
		repo:   r,
		logger: logger,
	}
}

func (uc *offersUseCase) GetOffers(ctx context.Context, filter domain.OffersFilter) ([]domain.Offer, error) {
	// Простейшая валидация
	if filter.ComponentID == "" {
		return nil, errors.New("componentId is required")
	}

	// Логика сортировки и т.д. можно уточнить
	// ... например, проверяем, что sortParam входит в допустимый список

	// Вызываем репозиторий
	offers, err := uc.repo.FetchOffers(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Можно добавить дополнительную логику (например, фильтр по цене и т.п.)
	return offers, nil
}

// ProcessUpdateEvent — вызывается Consumer’ом, когда приходит сообщение в очередь
func (uc *offersUseCase) ProcessUpdateEvent(ctx context.Context, evt domain.UpdateEvent) error {
	if evt.ShopID == "" {
		return errors.New("no shopId provided")
	}

	uc.logger.Printf("Processing update event for shopId=%s, action=%s", evt.ShopID, evt.Action)
	// Логика: обновить каталог, сходить в внешний API магазина, спарсить CSV/JSON, сохранить в offers
	// ...
	return nil
}
