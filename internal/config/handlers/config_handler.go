package handlers

import (
	"StartupPCConfigurator/internal/domain"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"StartupPCConfigurator/internal/config/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// / Создаём структуру, отражающую тело запроса в /config/newconfig.
// Это может соответствовать вашей OpenAPI (CreateConfigRequest).
type CreateConfigRequest struct {
	Name       string         `json:"name" binding:"required"`
	Components []ComponentRef `json:"components" binding:"required"`
}

type ComponentRef struct {
	Category string `json:"category"`
	Name     string `json:"name"`
}

// CreateConfig обрабатывает POST /config/newconfig
func (h *ConfigHandler) CreateConfig(c *gin.Context) {
	var req CreateConfigRequest
	// Считываем JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	uidAny, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no user id in context"})
		return
	}

	userID, ok := uidAny.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id type"})
		return
	}

	// Вызываем бизнес-логику
	config, err := h.service.CreateConfiguration(userID, req.Name, toDomainRefs(req.Components))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Успешный ответ: 201 + JSON с созданной конфигурацией
	c.JSON(http.StatusCreated, config)
}

// Создаём для PUT-запроса /config/newconfig/:configId
type UpdateConfigRequest struct {
	Name       string         `json:"name" binding:"required"` // или binding:"omitempty"
	Components []ComponentRef `json:"components" binding:"required"`
}

// UpdateConfig обрабатывает PUT /config/newconfig/:configId
func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	configId := c.Param("configId") // извлекаем из URL
	uidAny, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userId, ok := uidAny.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
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

// GetCompatibleComponents обрабатывает GET /config/compatible
func (h *ConfigHandler) GetCompatibleComponents(c *gin.Context) {
	filter := domain.CompatibilityFilter{
		Category:   c.Query("category"),
		CPUSocket:  c.Query("cpuSocket"),
		RAMType:    c.Query("memoryType"),
		FormFactor: c.Query("formFactor"),
	}

	// Пример распарсить числовые значения, если они есть
	if val := c.Query("gpuLengthMM"); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			filter.GPULengthMM = f
		}
	}
	if val := c.Query("coolerHeightMM"); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			filter.CoolerHeightMM = f
		}
	}
	if val := c.Query("powerRequired"); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			filter.PowerRequired = f
		}
	}

	comps, err := h.service.FetchCompatibleComponents(filter)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, comps)
}

// GetUserConfigs обрабатывает GET /config/userconf
func (h *ConfigHandler) GetUserConfigs(c *gin.Context) {
	raw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := raw.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id"})
		return
	}

	configs, err := h.service.FetchUserConfigurations(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, configs)
}

// DeleteConfig обрабатывает DELETE /config/newconfig/:configId
func (h *ConfigHandler) DeleteConfig(c *gin.Context) {
	configId := c.Param("configId")
	raw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := raw.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id"})
		return
	}

	err := h.service.DeleteConfiguration(userID, configId)
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

// ListUseCases обрабатывает GET /config/usecases
func (h *ConfigHandler) ListUseCases(c *gin.Context) {
	usecases, err := h.service.ListUseCases()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, usecases)
}

// GetUseCaseBuild обрабатывает GET /config/usecase/:name
func (h *ConfigHandler) GetUseCaseBuild(c *gin.Context) {
	name := c.Param("name")
	build, err := h.service.GetUseCaseBuild(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"components": build})
}

// GenerateUseCaseConfigsRequest — тело POST /config/usecase/:name/generate
type GenerateUseCaseConfigsRequest struct {
	Components []ComponentRef `json:"components"`
}

// GenerateUseCaseConfigs обрабатывает POST /config/usecase/:name/generate
func (h *ConfigHandler) GenerateUseCaseConfigs(c *gin.Context) {
	name := c.Param("name")

	var req GenerateUseCaseConfigsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	configs, err := h.service.GenerateUseCaseConfigs(name, toDomainRefs(req.Components))
	if err != nil {
		// различаем неизвестный сценарий и прочие ошибки
		if strings.HasPrefix(err.Error(), "unknown use case") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

// вспомогательная конвертация, как и для Create/Update
func toDomainRefs(input []ComponentRef) []domain.ComponentRef {
	var result []domain.ComponentRef
	for _, c := range input {
		result = append(result, domain.ComponentRef{
			Category: c.Category,
			Name:     c.Name,
		})
	}
	return result
}

// POST /config/generate
func (h *ConfigHandler) GenerateConfigs(c *gin.Context) {
	var req struct {
		Components []domain.ComponentRef `json:"components"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	configs, err := h.service.GenerateConfigurations(req.Components)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"configs": configs})
}
