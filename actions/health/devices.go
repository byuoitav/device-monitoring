package health

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/device-monitoring/localsystem"
)

const (
	healthyCommandID = "HealthCheck"
)

// GetDeviceAPIHealth .
func GetDeviceAPIHealth(ctx context.Context) (map[string]bool, *nerr.E) {
	log.L.Infof("Getting device api health")

	roomID, err := localsystem.RoomID()
	if err != nil {
		return nil, err.Addf("failed to get device api health")
	}

	devices, gerr := db.GetDB().GetDevicesByRoom(roomID)
	if gerr != nil {
		return nil, nerr.Translate(gerr).Addf("failed to get device api health")
	}

	healthy := make(map[string]bool)
	healthyMu := sync.Mutex{}
	wg := sync.WaitGroup{}

	for i := range devices {
		if devices[i].HasCommand(healthyCommandID) {
			wg.Add(1)

			go func(idx int) {
				defer wg.Done()
				h := isDeviceAPIHealthy(ctx, devices[idx])

				healthyMu.Lock()
				healthy[devices[idx].ID] = h
				healthyMu.Unlock()
			}(i)
		}
	}

	wg.Wait()
	return healthy, nil
}

func isDeviceAPIHealthy(ctx context.Context, device structs.Device) bool {
	// build the command
	address, err := device.BuildCommandURL(healthyCommandID)
	if err != nil {
		log.L.Warnf("unable to check if %s's API was healthy: %s", device.ID, err.Error())
		return false
	}

	// fill in the address
	address = strings.Replace(address, ":address", device.Address, 1)

	req, gerr := http.NewRequest("GET", address, nil)
	if gerr != nil {
		log.L.Warnf("unable to check if %s's API was healthy: %s", device.ID, gerr)
		return false
	}

	req = req.WithContext(ctx)
	resp, gerr := http.DefaultClient.Do(req)
	if gerr != nil {
		log.L.Warnf("unable to check if %s's API was healthy: %s", device.ID, gerr)
		return false
	}
	defer resp.Body.Close()

	bytes, gerr := ioutil.ReadAll(resp.Body)
	if gerr != nil {
		log.L.Warnf("unable to check if %s's API was healthy: %s", device.ID, gerr)
		return false
	}

	if resp.StatusCode/100 != 2 {
		log.L.Warnf("%s's API is unhealthy. response: %s", device.ID, bytes)
		return false
	}

	return true
}
