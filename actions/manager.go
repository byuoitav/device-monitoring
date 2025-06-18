package actions

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/byuoitav/common/db/couch"
	"github.com/byuoitav/device-monitoring/model"
	"github.com/byuoitav/shipwright/actions"
)

const (
	database = "device-monitoring"
)

var (
	managerOnce = sync.Once{}
	manager     *actions.ActionManager

	systemID = os.Getenv("SYSTEM_ID")
)

// ActionManager .
func ActionManager() *actions.ActionManager {
	managerOnce.Do(func() {
		manager = &actions.ActionManager{
			Config:      GetConfig(),
			Workers:     1000,
			EventStream: make(chan model.Event, 10000),
			EventCache:  "default",
		}
	})

	return manager
}

// GetConfig .
func GetConfig() *actions.ActionConfig {
	config := actions.ActionConfig{}
	db := couch.NewDB(os.Getenv("DB_ADDRESS"), os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"))

	if len(systemID) == 0 {
		slog.Error("failed to get action config: SYSTEM_ID not set")
		// If the system ID is not set, we cannot proceed with fetching the action config.
	}

	// get device specific jobs
	err := db.MakeRequest("GET", fmt.Sprintf("%v/%v", database, systemID), "", nil, &config)
	if err != nil {
		if _, ok := err.(*couch.NotFound); ok {
		} else if _, ok := err.(couch.NotFound); ok {
		} else {
			slog.Error("uable to get device monitoring actions", slog.String("error", err.Error()))
			// If the system ID is not set, we cannot proceed with fetching the action config.
		}
	} else {
		return &config
	}

	// get the current device's type
	dev, err := db.GetDevice(systemID)
	if err != nil {
		slog.Error("unable to get device monitoring actions: unable to get device type",
			slog.String("system_id", systemID),
			slog.String("error", err.Error()),
		)
	}

	// get the default actions for this device type
	err = db.MakeRequest("GET", fmt.Sprintf("%v/%v", database, dev.Type.ID), "", nil, &config)
	if err != nil {
		slog.Error("unable to get device monitoring actions for device type '%s'",
			slog.String("device_type", dev.Type.ID),
			slog.String("error", err.Error()),
		)
	}

	return &config
}
