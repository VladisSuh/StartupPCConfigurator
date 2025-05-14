package domain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// User — основная модель пользователя
type User struct {
	ID                    uuid.UUID `json:"id"`
	Email                 string    `json:"email"`
	EmailVerified         bool      `json:"email_verified"`
	IsSuperuser           bool      `json:"is_superuser"` // NEW: флаг суперпользователя
	VerificationCode      string    `json:"-"`
	PasswordHash          string    `json:"-"` // не выводится в JSON
	Name                  string    `json:"name"`
	CreatedAt             time.Time `json:"created_at"`
	RefreshToken          string    `json:"-"`
	RefreshTokenExpiresAt time.Time `json:"-"`
	ResetToken            string    `json:"-"`
	ResetTokenExpiresAt   time.Time `json:"-"`
}

// Token — структура для передачи JWT
type Token struct {
	AccessToken      string    `json:"access_token"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshToken     string    `json:"refresh_token"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

// Структура для таблицы components
type Component struct {
	ID        int             `gorm:"primaryKey" json:"id"`
	Name      string          `json:"name"`
	Category  string          `json:"category"`
	Brand     string          `json:"brand,omitempty"`
	Specs     json.RawMessage `gorm:"type:jsonb"       json:"specs"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// Это то, что мы получаем в CreateConfigRequest/UpdateConfigRequest
type ComponentRef struct {
	Category string `json:"category"`
	Name     string `json:"name"`
}

// Структура «Конфигурация» (сборка)
type Configuration struct {
	ID         int            `db:"id"`
	UserID     string         `db:"user_id"`
	Name       string         `db:"name"`
	Components []ComponentRef // не обязательно хранить здесь, т.к. в БД связь через configuration_components
	CreatedAt  time.Time      `db:"created_at"`
	UpdatedAt  time.Time      `db:"updated_at"`
	OwnerID    uuid.UUID
}

var (
	ErrConfigNotFound = errors.New("configuration not found")
	ErrForbidden      = errors.New("forbidden")
)

type Offer struct {
	ID           int64   `json:"-"`
	ComponentID  string  `json:"componentId"`
	ShopID       int64   `json:"shopId"`
	ShopCode     string  `json:"shopCode"`
	ShopName     string  `json:"shopName"`
	Price        float64 `json:"price"`
	Currency     string  `json:"currency"`
	Availability string  `json:"availability"`
	URL          string  `json:"url"`
	FetchedAt    string  `json:"fetchedAt"`
}

// OffersFilter описывает фильтры/параметры для запроса
type OffersFilter struct {
	ComponentID string
	Sort        string // "priceAsc" | "priceDesc" | ...
}

type UpdateEvent struct {
	ShopID   string `json:"shopId"`
	Action   string `json:"action"`
	Metadata string `json:"metadata"`
}

type CompatibilityFilter struct {
	Category      string // например "motherboard"
	NameComponent string
	Specs         map[string]interface{}
}

// UseCase — описание сценария сборки
type UseCase struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Notification struct {
	ID          uuid.UUID `db:"id" json:"id"`
	UserID      uuid.UUID `db:"user_id" json:"userId"`
	ComponentID string    `db:"component_id" json:"componentId"`
	ShopID      int64     `db:"shop_id" json:"shopId"`
	OldPrice    float64   `db:"old_price" json:"oldPrice"`
	NewPrice    float64   `db:"new_price" json:"newPrice"`
	IsRead      bool      `db:"is_read" json:"isRead"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
}
