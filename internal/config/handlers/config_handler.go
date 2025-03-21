package handlers

import (
	"StartupPCConfigurator/internal/domain"
	"errors"
	"net/http"

	"StartupPCConfigurator/internal/config/usecase"
	"github.com/gin-gonic/gin"
)

// / Создаём структуру, отражающую тело запроса в /config/newconfig.
// Это может соответствовать вашей OpenAPI (CreateConfigRequest).
type CreateConfigRequest struct {
	Name       string         `json:"name" binding:"required"`
	Components []ComponentRef `json:"components" binding:"required"`
}

type ComponentRef struct {
	Category    string `json:"category"`
	ComponentID string `json:"componentId"`
}

// CreateConfig обрабатывает POST /config/newconfig
func (h *ConfigHandler) CreateConfig(c *gin.Context) {
	var req CreateConfigRequest
	// Считываем JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Предположим, userId берётся из контекста после аутентификации (middleware).
	// Если у вас нет авторизации, можете захардкодить или передавать временно в заголовке.
	userId := c.GetString("userId")
	if userId == "" {
		// Если нужен userId, а его нет, значит 401
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no user id in context"})
		return
	}

	// Вызываем бизнес-логику
	config, err := h.service.CreateConfiguration(userId, req.Name, toDomainRefs(req.Components))
	if err != nil {
		// Например, если вернулась ошибка валидации (пустое имя, нет компонентов)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Успешный ответ: 201 + JSON с созданной конфигурацией
	c.JSON(http.StatusCreated, config)
}

// Вспомогательная функция, если в вашем сервисе ожидаются []repository.ComponentRef
// а здесь у вас локальные типы. Можно конвертировать.
func toDomainRefs(input []ComponentRef) []domain.ComponentRef {
	var result []domain.ComponentRef
	for _, c := range input {
		result = append(result, domain.ComponentRef{
			Category:    c.Category,
			ComponentID: c.ComponentID,
		})
	}
	return result
}

// Создаём для PUT-запроса /config/newconfig/:configId
type UpdateConfigRequest struct {
	Name       string         `json:"name" binding:"required"` // или binding:"omitempty"
	Components []ComponentRef `json:"components" binding:"required"`
}

// UpdateConfig обрабатывает PUT /config/newconfig/:configId
func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	configId := c.Param("configId") // извлекаем из URL
	userId := c.GetString("userId") // или временный "stubUser123" если нет авторизации
	if userId == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Парсим JSON
	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Вызываем сервис (бизнес-логика)
	updatedConfig, err := h.service.UpdateConfiguration(
		userId,
		configId,
		req.Name,
		toDomainRefs(req.Components), // переводим локальные ComponentRef -> domain.ComponentRef
	)
	if err != nil {
		// Примерный разбор ошибок
		switch {
		case errors.Is(err, domain.ErrConfigNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "configuration not found"})
		case errors.Is(err, domain.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	// Успешно обновлено
	c.JSON(http.StatusOK, updatedConfig)
}

// GetUserConfigs обрабатывает GET /config/userconf
func (h *ConfigHandler) GetUserConfigs(c *gin.Context) {
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	configs, err := h.service.FetchUserConfigurations(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, configs)
}

// DeleteConfig обрабатывает DELETE /config/newconfig/:configId
func (h *ConfigHandler) DeleteConfig(c *gin.Context) {
	configId := c.Param("configId")
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err := h.service.DeleteConfiguration(userId, configId)
	if err != nil {
		if err == usecase.ErrConfigNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "configuration not found"})
		} else if err == usecase.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent) // 204
}
