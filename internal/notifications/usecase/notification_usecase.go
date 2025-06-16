package usecase

import (
	"context"
	"time"

	"StartupPCConfigurator/internal/aggregator/usecase"
	"StartupPCConfigurator/internal/domain"
	"StartupPCConfigurator/internal/notifications/repository"
	"github.com/google/uuid"
	"log"
)

// NotificationUseCase описывает бизнес-логику уведомлений
type NotificationUseCase interface {
	// Обработка события изменения цены
	HandlePriceChange(ctx context.Context, msg usecase.PriceChangedMsg) error

	// Получение числа непрочитанных уведомлений для пользователя
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)

	// Список уведомлений для пользователя
	ListNotifications(ctx context.Context, userID uuid.UUID, page, pageSize int) (domain.PagedNotifications, error)

	// Отметить уведомление как прочитанное
	MarkAsRead(ctx context.Context, userID, notifID uuid.UUID) error

	Subscribe(ctx context.Context, userID uuid.UUID, componentID string) error

	Unsubscribe(ctx context.Context, userID uuid.UUID, componentID string) error

	CheckSubscribed(ctx context.Context,
		userID uuid.UUID, ids []string) (map[string]bool, error)
}

// notificationUseCase реализует NotificationUseCase
type notificationUseCase struct {
	repo   repository.NotificationRepository
	cache  repository.NotificationCache
	logger *log.Logger
}

// NewNotificationUseCase конструктор для NotificationUseCase
func NewNotificationUseCase(
	repo repository.NotificationRepository,
	cache repository.NotificationCache,
	logger *log.Logger,
) NotificationUseCase {
	return &notificationUseCase{repo: repo, cache: cache, logger: logger}
}

// HandlePriceChange сохраняет уведомления о смене цены и увеличивает кеш-счётчик
func (uc *notificationUseCase) HandlePriceChange(ctx context.Context, msg usecase.PriceChangedMsg) error {
	users, err := uc.repo.GetSubscribers(ctx, msg.ComponentID)
	if err != nil {
		uc.logger.Printf("GetSubscribers error: %v", err)
		return err
	}
	for _, userID := range users {
		notif := domain.Notification{
			ID:          uuid.New(),
			UserID:      userID,
			ComponentID: msg.ComponentID,
			ShopID:      msg.ShopID,
			OldPrice:    msg.OldPrice,
			NewPrice:    msg.NewPrice,
			IsRead:      false,
			CreatedAt:   time.Now(),
		}
		if err := uc.repo.CreateNotification(ctx, notif); err != nil {
			uc.logger.Printf("CreateNotification error: %v", err)
			continue
		}
		if err := uc.cache.IncrementUnread(ctx, userID); err != nil {
			uc.logger.Printf("IncrementUnread error: %v", err)
		}
	}
	return nil
}

// GetUnreadCount возвращает количество непрочитанных уведомлений
func (uc *notificationUseCase) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := uc.cache.GetUnreadCount(ctx, userID)
	if err != nil {
		uc.logger.Printf("GetUnreadCount error: %v", err)
		return 0, err
	}
	return count, nil
}

// ListNotifications возвращает список уведомлений из БД
func (uc *notificationUseCase) ListNotifications(ctx context.Context, userID uuid.UUID, page, pageSize int) (domain.PagedNotifications, error) {
	rows, total, err := uc.repo.ListNotifications(ctx, userID, page, pageSize)
	if err != nil {
		uc.logger.Printf("ListNotifications error: %v", err)
		return domain.PagedNotifications{}, err
	}

	// Собираем PagedNotifications
	return domain.PagedNotifications{
		Items:    rows,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// MarkAsRead помечает уведомление как прочитанное
func (uc *notificationUseCase) MarkAsRead(ctx context.Context, userID, notifID uuid.UUID) error {
	if err := uc.repo.MarkAsRead(ctx, userID, notifID); err != nil {
		uc.logger.Printf("MarkAsRead error: %v", err)
		return err
	}
	if err := uc.cache.ResetUnread(ctx, userID); err != nil {
		uc.logger.Printf("ResetUnread error: %v", err)
	}
	return nil
}

func (uc *notificationUseCase) Subscribe(ctx context.Context, userID uuid.UUID, componentID string) error {
	return uc.repo.Subscribe(ctx, userID, componentID)
}
func (uc *notificationUseCase) Unsubscribe(ctx context.Context, userID uuid.UUID, componentID string) error {
	return uc.repo.Unsubscribe(ctx, userID, componentID)
}

func (uc *notificationUseCase) CheckSubscribed(
	ctx context.Context, userID uuid.UUID, ids []string,
) (map[string]bool, error) {
	return uc.repo.SubscribedMap(ctx, userID, ids)
}
