package domain

import "time"

// User — основная модель пользователя
type User struct {
	ID                    uint      `json:"id"`
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
	Category    string `json:"category"`
	ComponentID string `json:"componentId"`
}

// Структура «Конфигурация» (сборка)
type Configuration struct {
	ID         int            `db:"id"`
	UserID     string         `db:"user_id"`
	Name       string         `db:"name"`
	Components []ComponentRef // не обязательно хранить здесь, т.к. в БД связь через configuration_components
	CreatedAt  time.Time      `db:"created_at"`
	UpdatedAt  time.Time      `db:"updated_at"`
}
