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

// TODO handle gated devices

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

	/*
		hasCommand := func(device structs.Device, cmdID string) bool {
			for i := range device.Type.Commands {
				if device.Type.Commands[i].ID == cmdID {
					return true
				}
			}

			return false
		}
	*/

	for i := range devices {
		if len(devices[i].GetCommandByID(healthyCommandID).ID) > 0 {
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
	address := device.GetCommandByID(healthyCommandID).BuildCommandAddress()
	if len(address) == 0 {
		log.L.Warnf("command '%s' does not exist on %s", healthyCommandID, device.ID)
		return true // we'll just assume that it's healthy if we can't check it
	}

	// fill in the address
	address = strings.Replace(address, ":address", device.Address, 1)

	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		log.L.Warnf("unable to check if %s's API was healthy: %s", device.ID, err)
		return false
	}

	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.L.Warnf("unable to check if %s's API was healthy: %s", device.ID, err)
		return false
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.L.Warnf("unable to check if %s's API was healthy: %s", device.ID, err)
		return false
	}

	if resp.StatusCode/100 != 2 {
		log.L.Warnf("%s's API is unhealthy. response: %s", device.ID, bytes)
		return false
	}

	return true
}
