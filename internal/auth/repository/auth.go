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
	Update(user *domain.User) error
	FindByRefreshToken(refreshToken string) (*domain.User, error)
	FindByResetToken(resetToken string) (*domain.User, error)
}

// userRepositoryPostgres — реализация репозитория для Postgres
type userRepositoryPostgres struct {
	db *sql.DB
}

// NewUserRepository возвращает реализацию UserRepository
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepositoryPostgres{db: db}
}

/*
Create создает в базе данных нового пользователя, надо добавить обработку ошибки, когда неправильно введен пароль/логин или что-то такое

уникальные почты 100%
*/
func (r *userRepositoryPostgres) Create(user *domain.User) error {
	query := "INSERT INTO users(email, password_hash, created_at, name) VALUES($1, $2, $3, $4) RETURNING id"
	return r.db.QueryRow(query, user.Email, user.PasswordHash, user.CreatedAt, user.Name).Scan(&user.ID)
}

// FindByEmail поиск пользователя по email
func (r *userRepositoryPostgres) FindByEmail(email string) (*domain.User, error) {
	query := "SELECT id, email, password_hash, name, created_at FROM users WHERE email = $1"
	row := r.db.QueryRow(query, email)
	var user domain.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // пользователь не найден
		}
		return nil, err
	}
	return &user, nil
}

//добавить функцию для поиска по id

func (r *userRepositoryPostgres) FindByID(id uint) (*domain.User, error) {
	query := "SELECT id, email, password_hash, name, created_at FROM users WHERE id = $1"
	row := r.db.QueryRow(query, id)
	var user domain.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // пользователь не найден
		}
		return nil, err
	}
	return &user, nil
}

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
