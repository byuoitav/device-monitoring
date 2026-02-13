package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/byuoitav/device-monitoring/actions/activesignal"
	"github.com/byuoitav/device-monitoring/actions/hardwareinfo"
	"github.com/byuoitav/device-monitoring/actions/health"
	"github.com/byuoitav/device-monitoring/actions/ping"
	"github.com/byuoitav/device-monitoring/actions/roomstate"
	"github.com/byuoitav/device-monitoring/localsystem"
	"github.com/gin-gonic/gin"
)

// PingRoom pings all devices in the room with a 10s timeout.
func PingRoom(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	roomID, err := localsystem.RoomID()
	if err != nil {
		slog.Error("unable to get room ID", slog.Any("error", err))
		c.String(http.StatusInternalServerError, fmt.Sprintf("unable to ping devices: %v", err))
		return
	}

	results, err := ping.Room(ctx, roomID, ping.Config{
		Count: 3,
		Delay: 1 * time.Second,
	}, slog.Default())
	if err != nil {
		slog.Error("ping room failed", slog.Any("error", err))
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, results)
}

// RoomHealth returns the AV‑API health of the room.
func RoomHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	statuses, err := health.GetDeviceAPIHealth(ctx)
	if err != nil {
		slog.Error("failed to get device API health", slog.Any("error", err))
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statuses)
}

// RoomState returns the AV‑API state of the room.
func RoomState(c *gin.Context) {
	roomID, err := localsystem.RoomID()
	if err != nil {
		slog.Error("failed to get room ID", slog.Any("error", err))
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	state, err := roomstate.Get(c.Request.Context(), roomID)
	if err != nil {
		slog.Error("failed to get room state", slog.Any("error", err))
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, state)
}

// ActiveSignal returns the current active inputs in the room.
func ActiveSignal(c *gin.Context) {
	activeMap, err := activesignal.GetMap(c.Request.Context())
	if err != nil {
		slog.Error("failed to get active signals", slog.Any("error", err))
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, activeMap)
}

// DeviceHardwareInfo returns hardware info for all devices in the room.
func DeviceHardwareInfo(c *gin.Context) {
	info, err := hardwareinfo.RoomDevicesInfo(c.Request.Context())
	if err != nil {
		slog.Error("failed to get device hardware info", slog.Any("error", err))
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, info)
}
