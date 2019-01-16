package handlers

import (
	"fmt"
	"net/http"

	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/device-monitoring/jobs"
	"github.com/byuoitav/device-monitoring/jobs/gpio"
	"github.com/labstack/echo"
)

// GetDividerState returns the state of the dividers
func GetDividerState(ctx echo.Context) error {
	job := &gpio.DividerSensorJob{}
	s := jobs.RunJob(job, nil)

	switch v := s.(type) {
	case error:
		return ctx.String(http.StatusInternalServerError, v.Error())
	case *nerr.E:
		return ctx.String(http.StatusInternalServerError, v.Error())
	case []*gpio.Pin:
		resp := make(map[string][]string)
		for i := range v {
			if v[i].Connected {
				resp["connected"] = append(resp["connected"], v[i].Displays)
			} else {
				resp["disconnected"] = append(resp["disconnected"], v[i].Displays)
			}
		}

		return ctx.JSON(http.StatusOK, resp)
	default:
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("unexpected type from job: %v", v))
	}
}

// PresetForHostname returns the preset that a specific hostname should be on
func PresetForHostname(ctx echo.Context) error {
	hostname := ctx.Param("hostname")

	job := &gpio.DividerSensorJob{}
	s := jobs.RunJob(job, nil)

	switch v := s.(type) {
	case error:
		return ctx.String(http.StatusInternalServerError, v.Error())
	case *nerr.E:
		return ctx.String(http.StatusInternalServerError, v.Error())
	case []*gpio.Pin:
		if len(v) == 0 || len(v) > 1 {
			return ctx.String(http.StatusBadRequest, fmt.Sprintf("not supported in this room"))
		}

		// TODO should be ctx.String()
		return ctx.JSON(http.StatusOK, v[0].CurrentPreset(hostname))
	default:
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("unexpected type from job: %v", v))
	}
}
