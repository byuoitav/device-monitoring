package handlers

import (
	"fmt"
	"net/http"

	"github.com/byuoitav/device-monitoring/actions/gpio"
	"github.com/labstack/echo"
)

// GetDividerState returns the state of the dividers
func GetDividerState(ctx echo.Context) error {
	v := gpio.GetPins()

	resp := make(map[string][]string)
	for i := range v {
		if v[i].Connected {
			resp["connected"] = append(resp["connected"], v[i].BlueberryPresets)
		} else {
			resp["disconnected"] = append(resp["disconnected"], v[i].BlueberryPresets)
		}
	}

	return ctx.JSON(http.StatusOK, resp)
}

// PresetForHostname returns the preset that a specific hostname should be on
func PresetForHostname(ctx echo.Context) error {
	hostname := ctx.Param("hostname")

	v := gpio.GetPins()

	if len(v) == 0 || len(v) > 1 {
		return ctx.String(http.StatusBadRequest, fmt.Sprintf("not supported in this room"))
	}

	return ctx.String(http.StatusOK, v[0].CurrentPreset(hostname))
}
