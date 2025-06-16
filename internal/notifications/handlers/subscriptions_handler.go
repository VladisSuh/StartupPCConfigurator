package handlers

import (
	"net/http"
	"strings"

	"StartupPCConfigurator/internal/notifications/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SubHandler struct {
	uc usecase.NotificationUseCase
}

func NewSubHandler(r *gin.RouterGroup, uc usecase.NotificationUseCase) {
	h := &SubHandler{uc: uc}
	r.GET("/status", h.batchStatus)
}

// GET /subscriptions/status?ids=1,2,3
func (h *SubHandler) batchStatus(c *gin.Context) {
	raw := c.Query("ids")
	if raw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids query param required"})
		return
	}
	ids := strings.Split(raw, ",")
	if len(ids) > 200 { // защитимся от DoS
		c.JSON(http.StatusBadRequest, gin.H{"error": "too many ids (max 200)"})
		return
	}

	uidVal, _ := c.Get("user_id")
	userID := uidVal.(uuid.UUID)

	m, err := h.uc.CheckSubscribed(c.Request.Context(), userID, ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m) // напр.: { "1":true, "2":false }
}
