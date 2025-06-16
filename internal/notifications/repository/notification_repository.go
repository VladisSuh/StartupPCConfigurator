package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"StartupPCConfigurator/internal/domain"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// NotificationRepository defines DB operations for notifications
type NotificationRepository interface {
	GetSubscribers(ctx context.Context, componentID string) ([]uuid.UUID, error)
	CreateNotification(ctx context.Context, n domain.Notification) error
	ListNotifications(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]domain.NotificationResponse, int, error)
	MarkAsRead(ctx context.Context, userID, notifID uuid.UUID) error
	Subscribe(ctx context.Context, userID uuid.UUID, componentID string) error
	Unsubscribe(ctx context.Context, userID uuid.UUID, componentID string) error
	SubscribedMap(ctx context.Context,
		userID uuid.UUID,
		ids []string) (map[string]bool, error)
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
// ListNotifications возвращает с джойном по components страницу уведомлений и общее число записей
func (r *repoImpl) ListNotifications(
	ctx context.Context,
	userID uuid.UUID,
	page, pageSize int,
) ([]domain.NotificationResponse, int, error) {
	// 1) Считаем общее число уведомлений (без LIMIT/OFFSET)
	var total int
	if err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM notifications WHERE user_id = $1", userID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 2) Основной SELECT с JOIN-ом на components
	const query = `
    SELECT
        n.id,
        n.component_id,
        c.name           AS component_name,
        c.category       AS component_category,
        n.shop_id,
        n.old_price,
        n.new_price,
        n.is_read,
        n.created_at
    FROM notifications n
    JOIN components c ON c.id = n.component_id::integer
    WHERE n.user_id = $1
    ORDER BY n.created_at DESC
    LIMIT $2 OFFSET $3
    `
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, query,
		userID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []domain.NotificationResponse
	for rows.Next() {
		var nr domain.NotificationResponse
		// у нас структура NotificationResponse: ID, ComponentID, ComponentName, ComponentCategory, ShopID, OldPrice, NewPrice, IsRead, CreatedAt
		if err := rows.Scan(
			&nr.ID,
			&nr.ComponentID,
			&nr.ComponentName,
			&nr.ComponentCategory,
			&nr.ShopID,
			&nr.OldPrice,
			&nr.NewPrice,
			&nr.IsRead,
			&nr.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, nr)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return out, total, nil
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

func (r *repoImpl) Subscribe(ctx context.Context, userID uuid.UUID, componentID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO subscriptions(user_id,component_id) VALUES($1,$2)
     ON CONFLICT DO NOTHING`,
		userID, componentID)
	return err
}

func (r *repoImpl) Unsubscribe(ctx context.Context, userID uuid.UUID, componentID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM subscriptions WHERE user_id=$1 AND component_id=$2`,
		userID, componentID,
	)
	return err
}

func (r *repoImpl) SubscribedMap(
	ctx context.Context, userID uuid.UUID, ids []string,
) (map[string]bool, error) {

	// 1. Если список пуст – сразу вернуть пустой map
	if len(ids) == 0 {
		return map[string]bool{}, nil
	}

	// 2. pg-array из []string → {"id1","id2"}
	//    pq.Array() не пригодится: нужен text[]
	arrayStr := "{" + strings.Join(ids, ",") + "}"

	const q = `
        SELECT component_id
        FROM   subscriptions
        WHERE  user_id = $1
          AND  component_id = ANY($2::text[])
    `
	rows, err := r.db.QueryContext(ctx, q, userID, arrayStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 3. default=false для всех; true - для найденных
	res := make(map[string]bool, len(ids))
	for _, id := range ids {
		res[id] = false
	}

	for rows.Next() {
		var cID string
		if err := rows.Scan(&cID); err != nil {
			return nil, err
		}
		res[cID] = true
	}
	return res, rows.Err()
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
