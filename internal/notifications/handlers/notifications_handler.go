package handlers

import (
	"log"
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
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID type"})
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
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID type"})
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
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID type"})
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

// POST /subscriptions
func (h *Handler) Subscribe(c *gin.Context) {
	log.Println(">> incoming Authorization:", c.GetHeader("Authorization"))
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	var body struct {
		ComponentID string `json:"componentId"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID type"})
		return
	}
	if err := h.uc.Subscribe(c.Request.Context(), userID, body.ComponentID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"componentId": body.ComponentID,
		"subscribed":  true,
	})
}

// DELETE /subscriptions/:componentId
func (h *Handler) Unsubscribe(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	compID := c.Param("componentId")
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID type"})
		return
	}
	if err := h.uc.Unsubscribe(c.Request.Context(), userID, compID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.Status(204)
}
