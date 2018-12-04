package handlers

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/device-monitoring/dmdb"
	"github.com/byuoitav/device-monitoring/jobs"
	"github.com/byuoitav/device-monitoring/localsystem"
	"github.com/labstack/echo"
)

const (
	maintKey = "maintenance"
)

func init() {
	val, err := isInMaintMode()
	if err != nil {
		log.L.Fatalf("failed to get initial maintenance mode value: %s", err)
	}

	log.L.Infof("Maintenance mode is set to %v", val)
}

// IsInMaintMode returns whether or not the ui is in maintenance mode
func IsInMaintMode(ctx echo.Context) error {
	val, err := isInMaintMode()
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("%v", err.Error()))
	}

	return ctx.String(http.StatusOK, fmt.Sprintf("%v", val))
}

// ToggleMaintMode swaps test mode to active/inactive
func ToggleMaintMode(ctx echo.Context) error {
	val, err := isInMaintMode()
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("%v", err.Error()))
	}

	val = !val
	log.L.Infof("Setting maintenance mode to %v", val)

	err = setMaintMode(val)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("%v", err.Error()))
	}

	jobs.Messenger().SendEvent(events.Event{
		GeneratingSystem: localsystem.MustSystemID(),
		Timestamp:        time.Now(),
		EventTags:        []string{},
		TargetDevice:     events.GenerateBasicDeviceInfo(localsystem.MustSystemID()),
		AffectedRoom:     events.GenerateBasicRoomInfo(localsystem.MustRoomID()),
		Key:              "in-maintenance-mode",
		Value:            fmt.Sprintf("%v", val),
	})

	return ctx.String(http.StatusOK, fmt.Sprintf("%v", val))
}

func isInMaintMode() (bool, *nerr.E) {
	val := false

	b, err := dmdb.Get(maintKey)
	if err != nil {
		return val, err.Addf("failed to get maintenance mode")
	}

	// it just hasn't been set yet, so we should set it to false
	if len(b) == 0 {
		err = setMaintMode(val)
		if err != nil {
			err.Addf("failed to get maintenance mode")
		}

		return val, err
	}

	dec := gob.NewDecoder(bytes.NewBuffer(b))
	gerr := dec.Decode(&val)
	if gerr != nil {
		return val, nerr.Translate(gerr).Addf("failed to get maintenance mode")
	}

	return val, nil
}

func setMaintMode(val bool) *nerr.E {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)

	gerr := enc.Encode(val)
	if gerr != nil {
		return nerr.Translate(gerr).Addf("failed to set maintenance mode")
	}

	err := dmdb.Put(maintKey, buf.Bytes())
	if err != nil {
		return err.Addf("failed to set maintenance mode")
	}

	return nil
}
