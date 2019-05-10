package health

import (
	"context"
	"fmt"
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
	// Healthy represents a healthy response
	Healthy = "healthy"

	healthyCommandID = "HealthCheck"
)

// GetDeviceAPIHealth .
func GetDeviceAPIHealth(ctx context.Context) (map[string]string, *nerr.E) {
	log.L.Infof("Getting device api health")

	roomID, err := localsystem.RoomID()
	if err != nil {
		return nil, err.Addf("failed to get device api health")
	}

	devices, gerr := db.GetDB().GetDevicesByRoom(roomID)
	if gerr != nil {
		return nil, nerr.Translate(gerr).Addf("failed to get device api health")
	}

	healthy := make(map[string]string)
	healthyMu := sync.Mutex{}
	wg := sync.WaitGroup{}

	for i := range devices {
		if devices[i].Address == "0.0.0.0" ||
			len(devices[i].Address) == 0 ||
			!devices[i].HasCommand(healthyCommandID) {
			continue
		}

		wg.Add(1)

		go func(idx int) {
			defer wg.Done()
			h := isDeviceAPIHealthy(ctx, devices[idx])

			healthyMu.Lock()
			healthy[devices[idx].ID] = h
			healthyMu.Unlock()
		}(i)
	}

	wg.Wait()
	return healthy, nil
}

func isDeviceAPIHealthy(ctx context.Context, device structs.Device) string {
	// build the command
	address, err := device.BuildCommandURL(healthyCommandID)
	if err != nil {
		return fmt.Sprintf("unable to check if API is healthy: %s", err.Error())
	}

	// fill in the address
	address = strings.Replace(address, ":address", device.Address, 1)

	req, gerr := http.NewRequest("GET", address, nil)
	if gerr != nil {
		return fmt.Sprintf("unable to check if API is healthy: %s", err.Error())
	}

	req = req.WithContext(ctx)
	resp, gerr := http.DefaultClient.Do(req)
	if gerr != nil {
		return fmt.Sprintf("unable to check if API is healthy: %s", gerr.Error())
	}
	defer resp.Body.Close()

	bytes, gerr := ioutil.ReadAll(resp.Body)
	if gerr != nil {
		return fmt.Sprintf("unable to check if API is healthy: %s", gerr.Error())
	}

	if resp.StatusCode/100 != 2 {
		return fmt.Sprintf("failed health check. response: %s", bytes)
	}

	return Healthy
}
