package handlers

import (
	"net/http"

	"github.com/byuoitav/device-monitoring/actions/health"
	"github.com/byuoitav/device-monitoring/localsystem"
	"github.com/gin-gonic/gin"
)

// GetDeviceHealth handles GET /api/v1/monitoring
// It retrieves the health status of devices in the current room.
func GetDeviceHealth(c *gin.Context) {
	roomID, err := localsystem.RoomID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get room ID"})
		return
	}
	c.Header("Content-Type", "application/json")
	results, err := health.GetRoomHealth(c.Request.Context(), roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get room health", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

// Plink handles GET /api/v1/monitoring/plink
// It checks if the service is reachable and returns a simple message. (plink is a common term for a simple connectivity check)
func Plink(c *gin.Context) {
	c.String(http.StatusOK, "Service is reachable")
}
