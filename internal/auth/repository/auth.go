package repository

import (
	"StartupPCConfigurator/internal/domain"
	"database/sql"

	"github.com/google/uuid"
)

// UserRepository описывает интерфейс для работы с пользователями
type UserRepository interface {
	Create(user *domain.User) error
	FindByEmail(email string) (*domain.User, error)
	FindByID(id uuid.UUID) (*domain.User, error)
	FindByRefreshToken(refreshToken string) (*domain.User, error)
	FindByResetToken(resetToken string) (*domain.User, error)
	Update(user *domain.User) error
	UpdatePassword(userID uuid.UUID, newPasswordHash string) error
	DeleteRefreshToken(userID uuid.UUID) error
	DeleteResetToken(userID uuid.UUID) error
	DeleteUser(userID uuid.UUID) error
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
	query := `
		INSERT INTO users (
			email, password_hash, created_at, name, verification_code
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	return r.db.QueryRow(
		query,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
		user.Name,
		user.VerificationCode,
	).Scan(&user.ID)
}

// FindByEmail ищет пользователя по email
func (r *userRepositoryPostgres) FindByEmail(email string) (*domain.User, error) {
	query := `
	SELECT id, email, password_hash, name, created_at,
	       refresh_token, refresh_token_expires_at,
	       reset_token, reset_token_expires_at,
	       verification_code, email_verified
	FROM users WHERE email=$1
`
	row := r.db.QueryRow(query, email)

	var user domain.User
	var refresh sql.NullString
	var refreshExp sql.NullTime
	var reset sql.NullString
	var resetExp sql.NullTime
	var verifCode sql.NullString
	var emailVerified bool

	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name,
		&user.CreatedAt, &refresh, &refreshExp,
		&reset, &resetExp,
		&verifCode, &emailVerified)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if refresh.Valid {
		user.RefreshToken = refresh.String
	}
	if refreshExp.Valid {
		user.RefreshTokenExpiresAt = refreshExp.Time
	}
	if reset.Valid {
		user.ResetToken = reset.String
	}
	if resetExp.Valid {
		user.ResetTokenExpiresAt = resetExp.Time
	}
	if verifCode.Valid {
		user.VerificationCode = verifCode.String
	}
	user.EmailVerified = emailVerified

	return &user, nil
}

// FindByID ищет пользователя по ID
func (r *userRepositoryPostgres) FindByID(id uuid.UUID) (*domain.User, error) {
	query := `
	  SELECT id, email, password_hash, name, created_at,
	         refresh_token, refresh_token_expires_at,
	         reset_token, reset_token_expires_at
	  FROM users WHERE id=$1`
	row := r.db.QueryRow(query, id)

	var user domain.User
	var refresh sql.NullString
	var refreshExp sql.NullTime
	var reset sql.NullString
	var resetExp sql.NullTime

	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name,
		&user.CreatedAt, &refresh, &refreshExp,
		&reset, &resetExp)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if refresh.Valid {
		user.RefreshToken = refresh.String
	}
	if refreshExp.Valid {
		user.RefreshTokenExpiresAt = refreshExp.Time
	}
	if reset.Valid {
		user.ResetToken = reset.String
	}
	if resetExp.Valid {
		user.ResetTokenExpiresAt = resetExp.Time
	}
	return &user, nil
}

// FindByRefreshToken ищет пользователя по refresh-токену
func (r *userRepositoryPostgres) FindByRefreshToken(rt string) (*domain.User, error) {
	query := `
	  SELECT id, email, password_hash, name, created_at,
	         refresh_token, refresh_token_expires_at
	  FROM users WHERE refresh_token=$1`
	row := r.db.QueryRow(query, rt)

	var user domain.User
	var refreshExp sql.NullTime

	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name,
		&user.CreatedAt, &user.RefreshToken, &refreshExp)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if refreshExp.Valid {
		user.RefreshTokenExpiresAt = refreshExp.Time
	}
	return &user, nil
}

// FindByResetToken ищет пользователя по reset-токену
func (r *userRepositoryPostgres) FindByResetToken(rt string) (*domain.User, error) {
	query := `
	  SELECT id, email, password_hash, name, created_at,
	         refresh_token, refresh_token_expires_at,
	         reset_token, reset_token_expires_at
	  FROM users WHERE reset_token=$1`
	row := r.db.QueryRow(query, rt)

	var user domain.User
	var refresh sql.NullString
	var refreshExp sql.NullTime
	var resetExp sql.NullTime

	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name,
		&user.CreatedAt, &refresh, &refreshExp,
		&user.ResetToken, &resetExp)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if refresh.Valid {
		user.RefreshToken = refresh.String
	}
	if refreshExp.Valid {
		user.RefreshTokenExpiresAt = refreshExp.Time
	}
	if resetExp.Valid {
		user.ResetTokenExpiresAt = resetExp.Time
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
func (r *userRepositoryPostgres) UpdatePassword(userID uuid.UUID, newPasswordHash string) error {
	query := "UPDATE users SET password_hash = $1 WHERE id = $2"
	_, err := r.db.Exec(query, newPasswordHash, userID)
	return err
}

// DeleteRefreshToken удаляет refresh-токен (logout)
func (r *userRepositoryPostgres) DeleteRefreshToken(userID uuid.UUID) error {
	query := "UPDATE users SET refresh_token = '', refresh_token_expires_at = NULL WHERE id = $1"
	_, err := r.db.Exec(query, userID)
	return err
}

// DeleteResetToken удаляет reset-токен
func (r *userRepositoryPostgres) DeleteResetToken(userID uuid.UUID) error {
	query := "UPDATE users SET reset_token = '', reset_token_expires_at = NULL WHERE id = $1"
	_, err := r.db.Exec(query, userID)
	return err
}

// DeleteUser удаляет пользователя
func (r *userRepositoryPostgres) DeleteUser(userID uuid.UUID) error {
	_, err := r.db.Exec("DELETE FROM users WHERE id = $1", userID)
	return err
}
