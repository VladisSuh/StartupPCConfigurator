package usecase

import (
	"StartupPCConfigurator/internal/config/repository"
	"errors"
	"fmt"
	"time"
)

var (
	ErrConfigNotFound = errors.New("configuration not found")
	ErrForbidden      = errors.New("not owner of configuration")
)

// Интерфейс, который будет использовать хендлер
type ConfigService interface {
	FetchComponents(category, search string) ([]repository.Component, error)
	CreateConfiguration(userId, name string, comps []repository.ComponentRef) (repository.Configuration, error)
	FetchUserConfigurations(userId string) ([]repository.Configuration, error)
	UpdateConfiguration(userId, configId string, name string, comps []repository.ComponentRef) (repository.Configuration, error)
	DeleteConfiguration(userId, configId string) error
}

// Реализация
type configService struct {
	repo repository.ConfigRepository
}

func NewConfigService(r repository.ConfigRepository) ConfigService {
	return &configService{repo: r}
}

func (s *configService) FetchComponents(category, search string) ([]repository.Component, error) {
	return s.repo.GetComponents(category, search)
}

func (s *configService) CreateConfiguration(userId, name string, comps []repository.ComponentRef) (repository.Configuration, error) {
	// тут можно проверить бизнес-логику (пустое имя? нет компонентов?)
	if name == "" {
		return repository.Configuration{}, errors.New("name is required")
	}
	if len(comps) == 0 {
		return repository.Configuration{}, errors.New("at least one component required")
	}

	// Можно проверить совместимость, если у нас есть правила:
	// isCompatible := checkCompatibility(comps)
	// if !isCompatible { return ..., ... }

	config, err := s.repo.CreateConfiguration(userId, name, comps)
	return config, err
}

func (s *configService) FetchUserConfigurations(userId string) ([]repository.Configuration, error) {
	return s.repo.GetUserConfigurations(userId)
}

func (s *configService) UpdateConfiguration(userId, configId string, name string, comps []repository.ComponentRef) (repository.Configuration, error) {
	// Проверить, что конфигурация принадлежит userId, что она существует
	// Проверить логику
	updated, err := s.repo.UpdateConfiguration(userId, configId, name, comps)
	if err != nil {
		return repository.Configuration{}, err
	}
	return updated, nil
}

func (s *configService) DeleteConfiguration(userId, configId string) error {
	// Проверить право на удаление
	return s.repo.DeleteConfiguration(userId, configId)
}
