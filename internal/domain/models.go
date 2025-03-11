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
