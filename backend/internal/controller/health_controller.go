package controller

import (
	"github.com/gin-gonic/gin"

	"github.com/enricojoe/dgram/backend/internal/util"
)

// HealthController exposes a liveness endpoint.
type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

// Register mounts the health routes on the given router group.
func (h *HealthController) Register(r *gin.RouterGroup) {
	r.GET("/health", h.health)
}

func (h *HealthController) health(c *gin.Context) {
	util.OK(c, gin.H{"status": "ok"})
}
