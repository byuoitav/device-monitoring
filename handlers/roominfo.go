package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/device-monitoring/actions/activesignal"
	"github.com/byuoitav/device-monitoring/actions/hardwareinfo"
	"github.com/byuoitav/device-monitoring/actions/health"
	"github.com/byuoitav/device-monitoring/actions/ping"
	"github.com/byuoitav/device-monitoring/actions/roomstate"
	"github.com/byuoitav/device-monitoring/localsystem"
	"github.com/labstack/echo"
)

// PingRoom pings all the devices for this room
func PingRoom(ectx echo.Context) error {
	ctx, cancel := context.WithTimeout(ectx.Request().Context(), 10*time.Second)
	defer cancel()

	roomID, err := localsystem.RoomID()
	if err != nil {
		return err.Addf("unable to ping devices")
	}

	results, err := ping.Room(ctx, roomID, ping.Config{
		Count: 3,
		Delay: 1 * time.Second,
	}, log.L)
	if err != nil {
		return ectx.String(http.StatusInternalServerError, err.String())
	}

	return ectx.JSON(http.StatusOK, results)
}

// RoomHealth .
func RoomHealth(ectx echo.Context) error {
	ctx, cancel := context.WithTimeout(ectx.Request().Context(), 10*time.Second)
	defer cancel()

	health, err := health.GetDeviceAPIHealth(ctx)
	if err != nil {
		return ectx.String(http.StatusInternalServerError, err.Error())
	}

	return ectx.JSON(http.StatusOK, health)
}

// RoomState returns the av-api state of the room
func RoomState(ectx echo.Context) error {
	roomID, err := localsystem.RoomID()
	if err != nil {
		return ectx.String(http.StatusInternalServerError, err.Error())
	}

	state, err := roomstate.Get(ectx.Request().Context(), roomID)
	if err != nil {
		return ectx.String(http.StatusInternalServerError, err.Error())
	}

	return ectx.JSON(http.StatusOK, state)
}

// ActiveSignal returns the active inputs in the room
func ActiveSignal(ectx echo.Context) error {
	active, err := activesignal.GetMap(ectx.Request().Context())
	if err != nil {
		return ectx.String(http.StatusInternalServerError, err.Error())
	}

	return ectx.JSON(http.StatusOK, active)
}

// DeviceHardwareInfo returns the hardware info for all devices in the room
func DeviceHardwareInfo(ectx echo.Context) error {
	info, err := hardwareinfo.RoomDevicesInfo(ectx.Request().Context())
	if err != nil {
		return ectx.String(http.StatusInternalServerError, err.Error())
	}

	return ectx.JSON(http.StatusOK, info)
}
