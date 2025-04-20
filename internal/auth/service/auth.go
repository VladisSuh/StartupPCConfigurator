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
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthService описывает методы аутентификации
type AuthService interface {
	Register(email, password, name string) (*domain.User, *domain.Token, error)
	Login(email, password string) (*domain.Token, error)
	Refresh(refreshToken string) (*domain.Token, error)
	GetUserByID(userID uuid.UUID) (*domain.User, error)
	ForgotPassword(email string) error
	ResetPassword(resetToken string, newPassword string) error
	VerifyEmail(email, verificationCode string) error
	Logout(userID uuid.UUID) error
	DeleteAccount(userID uuid.UUID) error
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

// Register - регистрация нового пользователя
func (s *authServiceImpl) Register(email, password, name string) (*domain.User, *domain.Token, error) {
	existingUser, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, nil, err
	}
	if existingUser != nil {
		return nil, nil, errors.New("user already exists")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}
	verificationCode, err := generateRandomToken(6)
	if err != nil {
		return nil, nil, err
	}

	user := &domain.User{
		Email:            email,
		PasswordHash:     string(hashed),
		Name:             name,
		CreatedAt:        time.Now(),
		VerificationCode: verificationCode,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, nil, err
	}

	if err := SendVerificationCodeToEmail(user.Email, verificationCode); err != nil {
		return nil, nil, err
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, nil, err
	}

	return user, token, nil
}

// Login - вход пользователя
func (s *authServiceImpl) Login(email, password string) (*domain.Token, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Генерируем `refresh_token`
	refreshToken, err := generateRandomToken(64)
	if err != nil {
		return nil, err
	}

	user.RefreshToken = refreshToken
	user.RefreshTokenExpiresAt = time.Now().Add(7 * 24 * time.Hour) // refresh живёт 7 дней

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	// Генерируем новый access-токен
	accessToken, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &domain.Token{
		AccessToken:  accessToken.AccessToken,
		ExpiresAt:    accessToken.ExpiresAt,
		RefreshToken: refreshToken,
	}, nil
}

// generateToken - создает JWT
func (s *authServiceImpl) generateToken(user *domain.User) (*domain.Token, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	return &domain.Token{
		AccessToken: tokenString,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}, nil
}

// Refresh - обновляет access_token
func (s *authServiceImpl) Refresh(refreshToken string) (*domain.Token, error) {
	user, err := s.userRepo.FindByRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}
	if user == nil || time.Now().After(user.RefreshTokenExpiresAt) {
		return nil, errors.New("refresh token expired or invalid")
	}

	newAccessToken, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := generateRandomToken(64)
	if err != nil {
		return nil, err
	}

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

// ForgotPassword - отправляет письмо со сбросом пароля
func (s *authServiceImpl) ForgotPassword(email string) error {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}

	resetToken, err := generateRandomToken(32)
	if err != nil {
		return err
	}

	user.ResetToken = resetToken
	user.ResetTokenExpiresAt = time.Now().Add(1 * time.Hour)

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	resetLink := "http://localhost:8080/auth/reset_password?token=" + resetToken
	return SendResetLinkToEmail(user.Email, resetLink)
}

// SendResetLinkToEmail - отправляет письмо на email
func SendResetLinkToEmail(email, resetLink string) error {
	from := "yourconfigurator@gmail.com"
	password := "gzrn jglq tcjq szon"
	smtpServer := "smtp.gmail.com"
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

func SendVerificationCodeToEmail(email, code string) error {
	from := "yourconfigurator@gmail.com"
	password := "gzrn jglq tcjq szon"
	smtpServer := "smtp.gmail.com"
	smtpPort := "587"

	to := []string{email}
	subject := "Подтверждение регистрации"
	body := fmt.Sprintf("Ваш код подтверждения email: %s", code)
	message := []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body))

	auth := smtp.PlainAuth("", from, password, smtpServer)
	return smtp.SendMail(smtpServer+":"+smtpPort, auth, from, to, message)
}

// ResetPassword - устанавливает новый пароль
func (s *authServiceImpl) ResetPassword(resetToken string, newPassword string) error {
	user, err := s.userRepo.FindByResetToken(resetToken)
	if err != nil {
		return err
	}
	if user == nil || time.Now().After(user.ResetTokenExpiresAt) {
		return errors.New("invalid or expired reset token")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdatePassword(user.ID, string(hashed)); err != nil {
		return err
	}

	return s.userRepo.DeleteResetToken(user.ID)
}

// VerifyEmail - подтверждает email
func (s *authServiceImpl) VerifyEmail(email, verificationCode string) error {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	fmt.Println("From client:", verificationCode)
	fmt.Println("From DB:    ", user.VerificationCode)
	fmt.Println("len(client):", len(verificationCode))
	fmt.Println("len(db):    ", len(user.VerificationCode))

	if user.VerificationCode != verificationCode {
		return errors.New("invalid verification code")
	}

	user.EmailVerified = true
	user.VerificationCode = ""

	return s.userRepo.Update(user)
}

// generateRandomToken - генерация случайного токена
func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GetUserByID - получает пользователя по ID
func (s *authServiceImpl) GetUserByID(userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// Logout - выход из профиля, удаляет refresh-токен
func (s *authServiceImpl) Logout(userID uuid.UUID) error {
	return s.userRepo.DeleteRefreshToken(userID)
}

// DeleteAccount - удаление аккаунта
func (s *authServiceImpl) DeleteAccount(userID uuid.UUID) error {
	return s.userRepo.DeleteUser(userID)
}
