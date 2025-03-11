//Бизнес-логика

package service

import (
	"StartupPCConfigurator/internal/auth/repository"
	"StartupPCConfigurator/internal/domain"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/smtp"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// AuthService описывает методы аутентификации
type AuthService interface {
	Register(email, password, name string) (*domain.User, *domain.Token, error)
	Login(email, password string) (*domain.Token, error)
	Refresh(refreshToken string) (*domain.Token, error)
	GetUserByID(userID uint) (*domain.User, error)
	ForgotPassword(email string) error
	ResetPassword(resetToken string, newPassword string) error
	VerifyEmail(email, verificationCode string) error
}

// authServiceImpl — реализация AuthService
type authServiceImpl struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

// NewAuthService создает экземпляр AuthService
func NewAuthService(repo repository.UserRepository, secret string) AuthService {
	return &authServiceImpl{
		userRepo:  repo,
		jwtSecret: secret,
	}
}

// Register реализует регистрацию нового пользователя
func (s *authServiceImpl) Register(email, password, name string) (*domain.User, *domain.Token, error) {
	// Проверка на существование
	existingUser, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, nil, err
	}
	if existingUser != nil {
		return nil, nil, errors.New("user already exists")
	}

	// Хеширование пароля
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	//Создание пользователя
	user := &domain.User{
		Email:        email,
		PasswordHash: string(hashed),
		Name:         name,
		CreatedAt:    time.Now(),
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, nil, err
	}

	// Генерация токена
	token, err := s.generateToken(user)
	if err != nil {
		return nil, nil, err
	}

	return user, token, nil
}

// Login реализует вход пользователя
func (s *authServiceImpl) Login(email, password string) (*domain.Token, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Генерируем токены
	accessToken, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}
	refreshToken, err := generateRandomToken(64)
	if err != nil {
		return nil, err
	}

	// Сохраняем refreshToken в БД
	user.RefreshToken = refreshToken
	user.RefreshTokenExpiresAt = time.Now().Add(7 * 24 * time.Hour) // refresh живёт 7 дней
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	// Теперь возвращаем `refresh_token`
	return &domain.Token{
		AccessToken:  accessToken.AccessToken,
		ExpiresAt:    accessToken.ExpiresAt,
		RefreshToken: refreshToken,
	}, nil
}

// generateToken создает JWT для пользователя
func (s *authServiceImpl) generateToken(user *domain.User) (*domain.Token, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateRandomToken(64) // Генерируем refresh токен
	if err != nil {
		return nil, err
	}

	// Обновляем refresh-токен в БД
	user.RefreshToken = refreshToken
	user.RefreshTokenExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return &domain.Token{
		AccessToken:  tokenString,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		RefreshToken: refreshToken,
	}, nil
}

// Refresh - обновляет токен, необходимый, чтобы клиент каждлый раз не входил в аккаунт
func (s *authServiceImpl) Refresh(refreshToken string) (*domain.Token, error) {
	// Проверяем refresh-токен в БД
	user, err := s.userRepo.FindByRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}
	if user == nil || time.Now().After(user.RefreshTokenExpiresAt) {
		return nil, errors.New("refresh token expired or invalid")
	}

	// Генерируем новые токены
	newAccessToken, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}
	newRefreshToken, err := generateRandomToken(64)
	if err != nil {
		return nil, err
	}

	// Обновляем refresh-токен в БД (удаляем старый!)
	user.RefreshToken = newRefreshToken
	user.RefreshTokenExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return &domain.Token{
		AccessToken:  newAccessToken.AccessToken,
		ExpiresAt:    newAccessToken.ExpiresAt,
		RefreshToken: newRefreshToken,
	}, nil
}

// ForgotPassword — отправляет письмо со ссылкой или кодом для сброса пароля
func (s *authServiceImpl) ForgotPassword(email string) error {

	// Находим пользователя по email
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Генерируем токен
	resetToken, err := generateRandomToken(32)
	if err != nil {
		return err
	}

	// Генерация токена для сброса пароля
	user.ResetToken = resetToken
	user.ResetTokenExpiresAt = time.Now().Add(1 * time.Hour)

	// Сохраняем изменения в БД:
	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	resetLink := "https://yourfrontend.com/reset-password?token=" + resetToken
	// Отправляем письмо со ссылкой/токеном
	if err := sendResetLinkToEmail(user.Email, resetLink); err != nil {
		return err
	}

	return nil
}

// ResetPassword — устанавливает новый пароль по действительному reset-токену
func (s *authServiceImpl) ResetPassword(resetToken string, newPassword string) error {
	// Находим пользователя по resetToken
	user, err := s.userRepo.FindByResetToken(resetToken)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("invalid or expired reset token")
	}
	// Проверяем срок действия
	if time.Now().After(user.ResetTokenExpiresAt) {
		return errors.New("reset token is expired")
	}

	// Хешируем новый пароль
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Сохраняем новый пароль
	user.PasswordHash = string(hashed)

	// Удаляем `reset_token`, чтобы нельзя было использовать повторно
	user.ResetToken = ""
	user.ResetTokenExpiresAt = time.Time{}

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	return nil
}

// VerifyEmail — подтверждает email пользователя по коду
func (s *authServiceImpl) VerifyEmail(email, verificationCode string) error {

	//Находим в БД email
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Сравниваем код, например user.VerificationCode
	if user.VerificationCode != verificationCode {
		return errors.New("invalid verification code")
	}

	// Отмечаем, что email подтвержден
	user.EmailVerified = true
	user.VerificationCode = "" // можно очистить

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	return nil
}

// ------------- Вспомогательные функции --------------

// generateRandomToken - генерирует на основе библиотеки rand токен размером в 64 бита
func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// sendResetLinkToEmail - отправляет письмо на email с помощью
func sendResetLinkToEmail(email, resetLink string) error {
	from := "uconf@mail.ru"
	password := "yourconfigurate1"
	smtpServer := "smtp.mail.ru"
	smtpPort := "587" //по стандарту
	to := []string{email}
	subject := "Восстановление пароля"
	body := fmt.Sprintf("Для восстановление пароля, нажмите на ссылку: %s", resetLink)
	message := []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body))

	auth := smtp.PlainAuth("", from, password, smtpServer)
	// Проверяем, что smtpPort — строка, иначе конвертируем
	err := smtp.SendMail(smtpServer+":"+smtpPort, auth, from, to, message)
	if err != nil {
		return err
	}
	return nil
}

func (s *authServiceImpl) GetUserByID(userID uint) (*domain.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}
