package usecase

import (
	"context"
)

// ShopComponent — пара component_id + URL страницы в магазине
type ShopComponent struct {
	ComponentID string
	URL         string
}

// Repository — всё, что нужно UpdateUseCase
type Repository interface {
	ListShopComponents(ctx context.Context, shopID int64) ([]ShopComponent, error)
	UpsertOffer(ctx context.Context, compID string, shopID int64, price float64, availability, url string) error
	InsertPriceHistory(ctx context.Context, compID string, shopID int64, price float64) error
	UpdateJobStatus(ctx context.Context, jobID int64, status string, message interface{}) error
	BulkUpsertOffers(ctx context.Context, recs []ImportRecord) error
	GetOfferPrice(ctx context.Context, compID string, shopID int64) (float64, error)
}
