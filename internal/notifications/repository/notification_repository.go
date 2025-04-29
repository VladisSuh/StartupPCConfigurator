package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"StartupPCConfigurator/internal/domain"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// NotificationRepository defines DB operations for notifications
type NotificationRepository interface {
	GetSubscribers(ctx context.Context, componentID string) ([]uuid.UUID, error)
	CreateNotification(ctx context.Context, n domain.Notification) error
	ListNotifications(ctx context.Context, userID uuid.UUID) ([]domain.Notification, error)
	MarkAsRead(ctx context.Context, userID, notifID uuid.UUID) error
}

// repoImpl implements NotificationRepository using PostgreSQL
type repoImpl struct {
	db *sql.DB
}

// NewNotificationRepository constructs a new NotificationRepository
func NewNotificationRepository(db *sql.DB) NotificationRepository {
	return &repoImpl{db: db}
}

// GetSubscribers returns user IDs subscribed to given component
func (r *repoImpl) GetSubscribers(ctx context.Context, componentID string) ([]uuid.UUID, error) {
	const query = `
SELECT user_id
FROM subscriptions
WHERE component_id = $1
`
	rows, err := r.db.QueryContext(ctx, query, componentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []uuid.UUID
	for rows.Next() {
		var u uuid.UUID
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// CreateNotification inserts a new notification record
func (r *repoImpl) CreateNotification(ctx context.Context, n domain.Notification) error {
	const query = `
INSERT INTO notifications
  (id, user_id, component_id, shop_id, old_price, new_price, is_read, created_at)
VALUES
  ($1, $2, $3, $4, $5, $6, $7, $8)
`
	_, err := r.db.ExecContext(ctx, query,
		n.ID, n.UserID, n.ComponentID, n.ShopID,
		n.OldPrice, n.NewPrice, n.IsRead, n.CreatedAt,
	)
	return err
}

// ListNotifications fetches all notifications for a user
func (r *repoImpl) ListNotifications(ctx context.Context, userID uuid.UUID) ([]domain.Notification, error) {
	const query = `
SELECT id, user_id, component_id, shop_id, old_price, new_price, is_read, created_at
FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.ComponentID,
			&n.ShopID, &n.OldPrice, &n.NewPrice, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, n)
	}
	return result, rows.Err()
}

// MarkAsRead sets a notification's is_read flag to true
func (r *repoImpl) MarkAsRead(ctx context.Context, userID, notifID uuid.UUID) error {
	const query = `
UPDATE notifications
SET is_read = TRUE
WHERE id = $1 AND user_id = $2
`
	_, err := r.db.ExecContext(ctx, query, notifID, userID)
	return err
}

// NotificationCache defines Redis operations for unread counts
type NotificationCache interface {
	IncrementUnread(ctx context.Context, userID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	ResetUnread(ctx context.Context, userID uuid.UUID) error
}

// cacheImpl implements NotificationCache using Redis
type cacheImpl struct {
	client *redis.Client
}

// NewNotificationCache constructs NotificationCache with given Redis client
func NewNotificationCache(client *redis.Client) NotificationCache {
	return &cacheImpl{client: client}
}

// IncrementUnread increases unread counter for a user
func (c *cacheImpl) IncrementUnread(ctx context.Context, userID uuid.UUID) error {
	key := fmt.Sprintf("notifications:%s:unread", userID.String())
	return c.client.Incr(ctx, key).Err()
}

// GetUnreadCount retrieves the unread count for a user
func (c *cacheImpl) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	key := fmt.Sprintf("notifications:%s:unread", userID.String())
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	cnt, err := strconv.Atoi(val)
	return cnt, err
}

// ResetUnread deletes the unread counter for a user
func (c *cacheImpl) ResetUnread(ctx context.Context, userID uuid.UUID) error {
	key := fmt.Sprintf("notifications:%s:unread", userID.String())
	return c.client.Del(ctx, key).Err()
}
