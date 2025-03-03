package service

import (
	"StartupPCConfigurator/internal/auth/repository"
	"StartupPCConfigurator/internal/domain"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// AuthService описывает методы аутентификации
type AuthService interface {
	Register(email, password, name string) (*domain.User, *domain.Token, error)
	Login(email, password string) (*domain.Token, error)
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

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}
	return token, nil
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
	return &domain.Token{
		AccessToken: tokenString,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}, nil
}
