package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"StartupPCConfigurator/internal/aggregator/usecase"
	"StartupPCConfigurator/internal/domain"
)

type OffersHandler struct {
	usecase usecase.OffersUseCase
}

// Конструктор
func NewOffersHandler(uc usecase.OffersUseCase) *OffersHandler {
	return &OffersHandler{
		usecase: uc,
	}
}

// GetOffers обрабатывает GET /offers?componentId=xxx&sort=priceAsc
func (h *OffersHandler) GetOffers(c *gin.Context) {
	componentID := c.Query("componentId")
	sortParam := c.Query("sort") // e.g. "priceAsc", "priceDesc", etc.

	// Подготавливаем фильтр
	filter := domain.OffersFilter{
		ComponentID: componentID,
		Sort:        sortParam,
	}

	// Вызываем бизнес-логику
	offers, err := h.usecase.GetOffers(c.Request.Context(), filter)
	if err != nil {
		// Например, если componentID пустой, вернуть 400
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Успешный ответ
	c.JSON(http.StatusOK, offers)
}

// POST /offers/import
func (h *OffersHandler) UploadPriceList(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open file"})
		return
	}
	defer f.Close()

	// вызываем usecase
	if err := h.usecase.ImportPriceList(c.Request.Context(), f); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
