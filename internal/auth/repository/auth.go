package repository

import (
	"StartupPCConfigurator/internal/domain"
	"database/sql"
)

// UserRepository описывает интерфейс для работы с пользователями
type UserRepository interface {
	Create(user *domain.User) error
	FindByEmail(email string) (*domain.User, error)
	FindByID(id uint) (*domain.User, error)
	FindByRefreshToken(refreshToken string) (*domain.User, error)
	FindByResetToken(resetToken string) (*domain.User, error)
	Update(user *domain.User) error
	UpdatePassword(userID uint, newPasswordHash string) error
	DeleteRefreshToken(userID uint) error
	DeleteResetToken(userID uint) error
}

// userRepositoryPostgres — реализация репозитория для Postgres
type userRepositoryPostgres struct {
	db *sql.DB
}

// NewUserRepository возвращает реализацию UserRepository
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepositoryPostgres{db: db}
}

// Create создаёт нового пользователя в базе данных
func (r *userRepositoryPostgres) Create(user *domain.User) error {
	query := "INSERT INTO users (email, password_hash, created_at, name) VALUES ($1, $2, $3, $4) RETURNING id"
	return r.db.QueryRow(query, user.Email, user.PasswordHash, user.CreatedAt, user.Name).Scan(&user.ID)
}

// FindByEmail ищет пользователя по email
func (r *userRepositoryPostgres) FindByEmail(email string) (*domain.User, error) {
	query := "SELECT id, email, password_hash, name, created_at, refresh_token, refresh_token_expires_at, reset_token, reset_token_expires_at FROM users WHERE email = $1"
	row := r.db.QueryRow(query, email)
	var user domain.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt, &user.RefreshToken, &user.RefreshTokenExpiresAt, &user.ResetToken, &user.ResetTokenExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // пользователь не найден
		}
		return nil, err
	}
	return &user, nil
}

// FindByID ищет пользователя по ID
func (r *userRepositoryPostgres) FindByID(id uint) (*domain.User, error) {
	query := "SELECT id, email, password_hash, name, created_at, refresh_token, refresh_token_expires_at, reset_token, reset_token_expires_at FROM users WHERE id = $1"
	row := r.db.QueryRow(query, id)
	var user domain.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt, &user.RefreshToken, &user.RefreshTokenExpiresAt, &user.ResetToken, &user.ResetTokenExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // пользователь не найден
		}
		return nil, err
	}
	return &user, nil
}

// FindByRefreshToken ищет пользователя по refresh-токену
func (r *userRepositoryPostgres) FindByRefreshToken(refreshToken string) (*domain.User, error) {
	query := "SELECT id, email, password_hash, name, created_at, refresh_token, refresh_token_expires_at FROM users WHERE refresh_token = $1"
	row := r.db.QueryRow(query, refreshToken)
	var user domain.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt, &user.RefreshToken, &user.RefreshTokenExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // refresh-токен не найден
		}
		return nil, err
	}
	return &user, nil
}

// FindByResetToken ищет пользователя по reset-токену
func (r *userRepositoryPostgres) FindByResetToken(resetToken string) (*domain.User, error) {
	query := "SELECT id, email, password_hash, name, created_at, refresh_token, refresh_token_expires_at, reset_token, reset_token_expires_at FROM users WHERE reset_token = $1"
	row := r.db.QueryRow(query, resetToken)
	var user domain.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt, &user.RefreshToken, &user.RefreshTokenExpiresAt, &user.ResetToken, &user.ResetTokenExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // reset-токен не найден
		}
		return nil, err
	}
	return &user, nil
}

// Update обновляет пользователя в БД
func (r *userRepositoryPostgres) Update(user *domain.User) error {
	query := `UPDATE users SET 
		email = $1, 
		password_hash = $2, 
		name = $3, 
		refresh_token = $4, 
		refresh_token_expires_at = $5, 
		reset_token = $6, 
		reset_token_expires_at = $7 
		WHERE id = $8`
	_, err := r.db.Exec(query, user.Email, user.PasswordHash, user.Name, user.RefreshToken, user.RefreshTokenExpiresAt, user.ResetToken, user.ResetTokenExpiresAt, user.ID)
	return err
}

// UpdatePassword обновляет пароль пользователя
func (r *userRepositoryPostgres) UpdatePassword(userID uint, newPasswordHash string) error {
	query := "UPDATE users SET password_hash = $1 WHERE id = $2"
	_, err := r.db.Exec(query, newPasswordHash, userID)
	return err
}

// DeleteRefreshToken удаляет refresh-токен (logout)
func (r *userRepositoryPostgres) DeleteRefreshToken(userID uint) error {
	query := "UPDATE users SET refresh_token = '', refresh_token_expires_at = NULL WHERE id = $1"
	_, err := r.db.Exec(query, userID)
	return err
}

// DeleteResetToken удаляет reset-токен
func (r *userRepositoryPostgres) DeleteResetToken(userID uint) error {
	query := "UPDATE users SET reset_token = '', reset_token_expires_at = NULL WHERE id = $1"
	_, err := r.db.Exec(query, userID)
	return err
}
