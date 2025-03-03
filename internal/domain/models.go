package domain

import "time"

// User — основная модель пользователя
type User struct {
	ID           uint      `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // не выводится в JSON
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
}

// Token — структура для передачи JWT
type Token struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}
