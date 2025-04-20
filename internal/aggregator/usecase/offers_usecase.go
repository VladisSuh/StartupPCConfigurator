package usecase

import (
	"StartupPCConfigurator/internal/domain"
	"context"
	"errors"
)

// OffersRepository — только для GET /offers
type OffersRepository interface {
	FetchOffers(ctx context.Context, filter domain.OffersFilter) ([]domain.Offer, error)
}

type OffersUseCase interface {
	GetOffers(ctx context.Context, filter domain.OffersFilter) ([]domain.Offer, error)
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
