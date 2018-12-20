package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/device-monitoring/jobs"
	"github.com/byuoitav/device-monitoring/jobs/ask"
	"github.com/labstack/echo"
)

// GetRoom .
func GetRoom(context echo.Context) error {
	data := make(map[string]interface{})

	// pulse the room state
	stateJob := &ask.StateUpdateJob{}
	state := jobs.RunJob(stateJob, nil)

	switch v := state.(type) {
	case error:
		data["error"] = fmt.Sprintf("%v", v)
		return context.JSON(http.StatusInternalServerError, data)
	case *nerr.E:
		data["error"] = fmt.Sprintf("%v", v)
		return context.JSON(http.StatusInternalServerError, data)
	case base.PublicRoom:
		data["state"] = v
	case *base.PublicRoom:
		data["state"] = v
	default:
		data["error"] = fmt.Sprintf("unexpected type from mstatus job: %v", v)
		return context.JSON(http.StatusInternalServerError, data)
	}

	pingJob := &ask.PingJob{
		Count:    4,
		Interval: 1 * time.Second,
		Timeout:  2 * time.Second,
	}
	result := jobs.RunJob(pingJob, nil)

	switch v := result.(type) {
	case error:
		data["error"] = fmt.Sprintf("%v", v)
		return context.JSON(http.StatusInternalServerError, data)
	case *nerr.E:
		data["error"] = fmt.Sprintf("%v", v)
		return context.JSON(http.StatusInternalServerError, data)
	case ask.PingResult:
		data["ping-result"] = v
	case *ask.PingResult:
		data["ping-result"] = v
	default:
		data["error"] = fmt.Sprintf("unexpected type from mstatus job: %v", v)
		return context.JSON(http.StatusInternalServerError, data)
	}

	return context.JSON(http.StatusOK, data)
}

// RoomState returns the room state, but also pulses it around the room
func RoomState(context echo.Context) error {
	// pulse the room state
	job := &ask.StateUpdateJob{}
	state := jobs.RunJob(job, nil)

	switch v := state.(type) {
	case error:
		return context.String(http.StatusInternalServerError, v.Error())
	case *nerr.E:
		return context.String(http.StatusInternalServerError, v.Error())
	case base.PublicRoom:
		return context.JSON(http.StatusOK, v)
	case *base.PublicRoom:
		return context.JSON(http.StatusOK, v)
	default:
		return context.String(http.StatusInternalServerError, fmt.Sprintf("unexpected type from job: %v", v))
	}
}

// PingStatus .
func PingStatus(context echo.Context) error {
	job := &ask.PingJob{
		Count:    3,
		Interval: 500 * time.Millisecond,
		Timeout:  1500 * time.Millisecond,
	}
	result := jobs.RunJob(job, nil)

	switch v := result.(type) {
	case error:
		return context.String(http.StatusInternalServerError, v.Error())
	case *nerr.E:
		return context.String(http.StatusInternalServerError, v.Error())
	case ask.PingResult:
		return context.JSON(http.StatusOK, v)
	case *ask.PingResult:
		return context.JSON(http.StatusOK, v)
	default:
		return context.String(http.StatusInternalServerError, fmt.Sprintf("unexpected type from job: %v", v))
	}
}

// ActiveSignal returns the active inputs in the room
func ActiveSignal(context echo.Context) error {
	job := &ask.ActiveSignalJob{}
	active := jobs.RunJob(job, nil)

	switch v := active.(type) {
	case error:
		return context.String(http.StatusInternalServerError, v.Error())
	case *nerr.E:
		return context.String(http.StatusInternalServerError, v.Error())
	case map[string]bool:
		return context.JSON(http.StatusOK, v)
	default:
		return context.String(http.StatusInternalServerError, fmt.Sprintf("unexpected type from job: %v", v))
	}
}

// DeviceHardwareInfo returns the hardware info for all devices in the room
func DeviceHardwareInfo(context echo.Context) error {
	job := &ask.DeviceHardwareJob{}
	info := jobs.RunJob(job, nil)

	switch v := info.(type) {
	case error:
		return context.String(http.StatusInternalServerError, v.Error())
	case *nerr.E:
		return context.String(http.StatusInternalServerError, v.Error())
	case map[string]structs.HardwareInfo:
		return context.JSON(http.StatusOK, v)
	default:
		return context.String(http.StatusInternalServerError, fmt.Sprintf("unexpected type from job: %v", v))
	}
}
