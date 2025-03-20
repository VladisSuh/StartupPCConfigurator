package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"StartupPCConfigurator/internal/config/usecase"
)

// CreateConfig обрабатывает POST /config/newconfig
func (h *ConfigHandler) CreateConfig(c *gin.Context) {
	var req CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Предположим, userId мы берём из токена (например, в middleware),
	// или в данном примере захардкожен/передан в заголовке
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	config, err := h.service.CreateConfiguration(userId, req.Name, req.Components)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, config)
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

// UpdateConfig обрабатывает PUT /config/newconfig/:configId
func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	configId := c.Param("configId")
	userId := c.GetString("userId")
	if userId == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	updated, err := h.service.UpdateConfiguration(userId, configId, req.Name, req.Components)
	if err != nil {
		// примерный разбор ошибок
		if err == usecase.ErrConfigNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "configuration not found"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, updated)
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
