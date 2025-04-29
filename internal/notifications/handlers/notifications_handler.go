package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"StartupPCConfigurator/internal/notifications/usecase"
)

// Handler содержит методы для HTTP API уведомлений
type Handler struct {
	uc usecase.NotificationUseCase
}

// NewHandler создаёт новый экземпляр Notification Handler
func NewHandler(uc usecase.NotificationUseCase) *Handler {
	return &Handler{uc: uc}
}

// UnreadCount обрабатывает GET /notifications/count
func (h *Handler) UnreadCount(c *gin.Context) {
	// userId извлекаем из контекста, устанавливается в middleware
	userIDStr := c.GetString("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	cnt, err := h.uc.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread": cnt})
}

// List обрабатывает GET /notifications
func (h *Handler) List(c *gin.Context) {
	userIDStr := c.GetString("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	notifications, err := h.uc.ListNotifications(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Преобразуем domain.Notification в JSON-ответ
	resp := make([]gin.H, 0, len(notifications))
	for _, n := range notifications {
		rep := gin.H{
			"id":          n.ID,
			"componentId": n.ComponentID,
			"shopId":      n.ShopID,
			"oldPrice":    n.OldPrice,
			"newPrice":    n.NewPrice,
			"isRead":      n.IsRead,
			"createdAt":   n.CreatedAt.Format(time.RFC3339),
		}
		resp = append(resp, rep)
	}

	c.JSON(http.StatusOK, resp)
}

// MarkRead обрабатывает POST /notifications/:id/read
func (h *Handler) MarkRead(c *gin.Context) {
	userIDStr := c.GetString("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// парсим ID уведомления из параметра
	notifIDStr := c.Param("id")
	notifID, err := uuid.Parse(notifIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	err = h.uc.MarkAsRead(c.Request.Context(), userID, notifID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
