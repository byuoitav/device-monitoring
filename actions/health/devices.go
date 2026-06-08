package health

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/device-monitoring/couchdb"
	"github.com/byuoitav/device-monitoring/localsystem"
	"github.com/byuoitav/device-monitoring/model"
)

const (
	// Healthy represents a healthy response
	Healthy = "healthy"

	healthyCommandID = "HealthCheck"
	maxHealthBody    = 64 << 10
)

var deviceHealthHTTPClient = &http.Client{
	Timeout: 5 * time.Second,
}

// GetDeviceAPIHealth queries all devices in the current room and returns a map
// of deviceID -> health status string, or an error.
func GetDeviceAPIHealth(ctx context.Context) (map[string]string, error) {
	slog.Info("Getting device API health")

	roomID, err := localsystem.RoomID()
	if err != nil {
		return nil, fmt.Errorf("failed to get room ID: %w", err)
	}

	devices, err := couchdb.GetDevicesByRoom(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices in room %q: %w", roomID, err)
	}
	slog.Info("Devices fetched", slog.Int("count", len(devices)), slog.String("room_id", roomID))

	healthy := make(map[string]string, len(devices))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, dev := range devices {
		if len(dev.Address) == 0 ||
			strings.HasPrefix(dev.Address, "0.0.0.") ||
			!dev.HasCommand(healthyCommandID) {
			continue
		}

		wg.Add(1)
		go func(d model.Device) {
			defer wg.Done()
			status := isDeviceAPIHealthy(ctx, d)

			mu.Lock()
			healthy[d.ID] = status
			mu.Unlock()
		}(dev)
	}

	wg.Wait()
	return healthy, nil
}

func isDeviceAPIHealthy(ctx context.Context, device model.Device) string {
	// build the command
	address, err := device.BuildCommandURL(healthyCommandID)
	//slog.Debug("Building command URL", slog.String("device_id", device.ID), slog.String("address", address))
	if err != nil {
		return fmt.Sprintf("unable to check if API is healthy: %s", err.Error())
	}

	// fill in the address
	address = strings.Replace(address, ":address", device.Address, 1)

	req, gerr := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if gerr != nil {
		return fmt.Sprintf("unable to check if API is healthy: %s", gerr.Error())
	}

	resp, gerr := deviceHealthHTTPClient.Do(req)
	if gerr != nil {
		return fmt.Sprintf("unable to check if API is healthy: %s", gerr.Error())
	}
	defer resp.Body.Close()

<<<<<<< HEAD
	bytes, gerr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
=======
	bytes, gerr := io.ReadAll(io.LimitReader(resp.Body, maxHealthBody))
>>>>>>> 8f6d957 (adding context to health check requests with timeouts)
	if gerr != nil {
		return fmt.Sprintf("unable to check if API is healthy: %s", gerr.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("failed health check. response: %s", bytes)
	}

	return Healthy
}
