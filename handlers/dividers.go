package handlers

import (
	"fmt"
	"net/http"

	"github.com/byuoitav/device-monitoring/actions/gpio"
	"github.com/gin-gonic/gin"
)

// GetDividerState returns the state of the dividers
func GetDividerState(c *gin.Context) {
	pins := gpio.GetPins()

	resp := make(map[string][]string)
	for _, p := range pins {
		if p.Connected {
			resp["connected"] = append(resp["connected"], p.BlueberryPresets)
		} else {
			resp["disconnected"] = append(resp["disconnected"], p.BlueberryPresets)
		}
	}

	c.JSON(http.StatusOK, resp)
}

// PresetForHostname returns the preset that a specific hostname should be on
func PresetForHostname(c *gin.Context) {
	hostname := c.Param("hostname")

	pins := gpio.GetPins()

	if len(pins) != 1 {
		c.String(http.StatusBadRequest, fmt.Sprintf("not supported in this room"))
		return
	}
	preset := pins[0].CurrentPreset(hostname)
	c.String(http.StatusOK, preset)
}
