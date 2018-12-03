package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/device-monitoring/jobs"
	"github.com/byuoitav/device-monitoring/localsystem"
	"github.com/labstack/echo"
)

var inMaintMode bool

// IsInMaintMode returns whether or not the ui is in maintenance mode
func IsInMaintMode(ctx echo.Context) error {
	return ctx.String(http.StatusOK, fmt.Sprintf("%v", inMaintMode))
}

// ToggleMaintMode swaps test mode to active/inactive
func ToggleMaintMode(ctx echo.Context) error {
	inMaintMode = !inMaintMode
	log.L.Infof("Setting maintenance mode to %v", inMaintMode)

	jobs.Messenger().SendEvent(events.Event{
		GeneratingSystem: localsystem.MustSystemID(),
		Timestamp:        time.Now(),
		EventTags:        []string{},
		TargetDevice:     events.GenerateBasicDeviceInfo(localsystem.MustSystemID()),
		AffectedRoom:     events.GenerateBasicRoomInfo(localsystem.MustRoomID()),
		Key:              "in-maintenance-mode",
		Value:            fmt.Sprintf("%v", inMaintMode),
	})

	return ctx.String(http.StatusOK, fmt.Sprintf("%v", inMaintMode))
}
