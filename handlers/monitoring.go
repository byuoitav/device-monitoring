package handlers

import (
	"net/http"

	"github.com/byuoitav/device-monitoring/actions/health"
	"github.com/gin-gonic/gin"
)

// TODO:
// 1. Get all the Devices in the room
// 2. For each device, get the Device api health
// 3. For each device, check if the device is healthy
// 4. If the device is healthy, return healthy
// 5. If the device is not healthy, return the health of the device (error message)
// 6. Return the health of each device (json or something)

// GetDeviceHealth handles GET /api/v1/monitoring?room_id=<room_id>
func GetDeviceHealth(c *gin.Context) {
	roomID := c.Query("room_id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing room_id"})
		return
	}

	results, err := health.GetDeviceHealth(c.Request.Context(), roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": results})
}
