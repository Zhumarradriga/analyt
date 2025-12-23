package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"analytservice/internal/domain"
	"analytservice/internal/service"
)

type EventHandler struct {
	svc *service.AnalyticsService
}

func NewEventHandler(svc *service.AnalyticsService) *EventHandler {
	return &EventHandler{svc: svc}
}


func (h *EventHandler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		api.POST("/track", h.TrackEvent)
		api.GET("/health", h.Health)
	}
}

func (h *EventHandler) TrackEvent(c *gin.Context) {
	var event domain.Event
	
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.svc.Track(event)

	c.JSON(http.StatusAccepted, gin.H{"status": "queued"})
}

func (h *EventHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}