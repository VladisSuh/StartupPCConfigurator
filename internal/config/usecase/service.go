package usecase

import (
	"StartupPCConfigurator/internal/config/repository"
	"StartupPCConfigurator/internal/domain"
	"errors"
	_ "fmt"
	_ "time"
)

var (
	ErrConfigNotFound = errors.New("configuration not found")
	ErrForbidden      = errors.New("not owner of configuration")
)

// Интерфейс, который будет использовать хендлер
type ConfigService interface {
	FetchComponents(category, search string) ([]domain.Component, error)
	FetchCompatibleComponents(category, cpuSocket, memoryType string) ([]domain.Component, error)
	CreateConfiguration(userId, name string, comps []domain.ComponentRef) (domain.Configuration, error)
	FetchUserConfigurations(userId string) ([]domain.Configuration, error)
	UpdateConfiguration(userId, configId string, name string, comps []domain.ComponentRef) (domain.Configuration, error)
	DeleteConfiguration(userId, configId string) error
}

// Реализация
type configService struct {
	repo repository.ConfigRepository
}

func (s *configService) FetchCompatibleComponents(category, cpuSocket, memoryType string) ([]domain.Component, error) {
	// Можно реализовать логику совместимости на уровне бизнес-логики.
	// Например: если запрошена категория "motherboard" и указан cpuSocket,
	// фильтровать материнские платы по указанному сокету.
	// В качестве упрощённого варианта сразу перенаправим запрос в репозиторий.

	return s.repo.GetCompatibleComponents(category, cpuSocket, memoryType)
}

func NewConfigService(r repository.ConfigRepository) ConfigService {
	return &configService{repo: r}
}

func (s *configService) FetchComponents(category, search string) ([]domain.Component, error) {
	return s.repo.GetComponents(category, search)
}

func (s *configService) CreateConfiguration(userId, name string, comps []domain.ComponentRef) (domain.Configuration, error) {
	// тут можно проверить бизнес-логику (пустое имя? нет компонентов?)
	if name == "" {
		return domain.Configuration{}, errors.New("name is required")
	}
	if len(comps) == 0 {
		return domain.Configuration{}, errors.New("at least one component required")
	}

	// Можно проверить совместимость, если у нас есть правила:
	// isCompatible := checkCompatibility(comps)
	// if !isCompatible { return ..., ... }

	config, err := s.repo.CreateConfiguration(userId, name, comps)
	return config, err
}

func (s *configService) FetchUserConfigurations(userId string) ([]domain.Configuration, error) {
	return s.repo.GetUserConfigurations(userId)
}

func (s *configService) UpdateConfiguration(userId, configId string, name string, comps []domain.ComponentRef) (domain.Configuration, error) {
	// Проверить, что конфигурация принадлежит userId, что она существует
	// Проверить логику
	updated, err := s.repo.UpdateConfiguration(userId, configId, name, comps)
	if err != nil {
		if errors.Is(err, domain.ErrConfigNotFound) {
			return domain.Configuration{}, domain.ErrConfigNotFound
		} else if errors.Is(err, domain.ErrForbidden) {
			return domain.Configuration{}, domain.ErrForbidden
		}
		return domain.Configuration{}, err
	}
	return updated, nil
}

func (s *configService) DeleteConfiguration(userId, configId string) error {
	err := s.repo.DeleteConfiguration(userId, configId)
	if err != nil {
		if errors.Is(err, domain.ErrConfigNotFound) {
			return domain.ErrConfigNotFound
		} else if errors.Is(err, domain.ErrForbidden) {
			return domain.ErrForbidden
		}
		return err
	}
	return nil
}
