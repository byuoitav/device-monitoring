package actions

import (
	"fmt"
	"os"
	"sync"

	"github.com/byuoitav/common/db/couch"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/v2/events"
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
			EventStream: make(chan events.Event, 10000),
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
		log.L.Fatalf("failed to get action config: SYSTEM_ID not set")
	}

	// get device specific jobs
	err := db.MakeRequest("GET", fmt.Sprintf("%v/%v", database, systemID), "", nil, &config)
	if err != nil {
		if _, ok := err.(*couch.NotFound); ok {
		} else if _, ok := err.(couch.NotFound); ok {
		} else {
			log.L.Fatalf("unable to get device monitoring actions: %s", err)
		}
	} else {
		return &config
	}

	// get the current device's type
	dev, err := db.GetDevice(systemID)
	if err != nil {
		log.L.Fatalf("unable to get device monitoring actions: unable to get device type: %s", err)
	}

	// get the default actions for this device type
	err = db.MakeRequest("GET", fmt.Sprintf("%v/%v", database, dev.Type.ID), "", nil, &config)
	if err != nil {
		log.L.Fatalf("unable to get device monitoring actions for device type '%s': %s", dev.Type.ID, err)
	}

	return &config
}
