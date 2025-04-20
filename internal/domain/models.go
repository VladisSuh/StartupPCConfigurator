package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// User — основная модель пользователя
type User struct {
	ID                    uuid.UUID `json:"id"`
	Email                 string    `json:"email"`
	EmailVerified         bool      `json:"email_verified"`
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
	ID        int       `db:"id"` // или uuid, если хотите
	Name      string    `db:"name"`
	Category  string    `db:"category"`
	Brand     string    `db:"brand"`
	Specs     []byte    `db:"specs"` // можно хранить как JSON-сырые данные
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
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
	ShopID       string  `json:"shopId"`
	ShopCode     int64   `json:"shopCode"`
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
	Category       string  // например "motherboard"
	CPUSocket      string  // например "AM4"
	RAMType        string  // например "DDR4"
	FormFactor     string  // например "ATX"
	GPULengthMM    float64 // например 300.0
	CoolerHeightMM float64 // например 160.0
	PowerRequired  float64 // для GPU (в ваттах)
}
