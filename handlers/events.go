package handlers

import (
	"fmt"
	"net/http"

	"github.com/byuoitav/device-monitoring/messenger"
	"github.com/byuoitav/device-monitoring/model"
	"github.com/gin-gonic/gin"
)

// SendEvent injects an event into the event mesh
func SendEvent(c *gin.Context) {
	event := model.Event{}

	err := c.BindJSON(&event)
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("failed to bind event: %v", err))
		return
	}

	messenger.Get().SendEvent(model.ToCommonEvent(event))
	c.String(http.StatusOK, "event sent")
}
