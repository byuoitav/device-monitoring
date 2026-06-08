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
	"github.com/byuoitav/device-monitoring/model"
)

const (
	healthyStatus  = "healthy"
	healthCheckCmd = "HealthCheck"
)

var healthHTTPClient = &http.Client{Timeout: 10 * time.Second}

// HealthStatus is the JSON‑serializable result for one device.
type HealthStatus struct {
	DeviceID string `json:"device_id"`
	Status   string `json:"status"` // "healthy" or "error"
	// If Status is "healthy", no Error field is populated.
	Error string `json:"error,omitempty"` // populated if Status=="error"
}

// GetDeviceHealth looks up all devices in the room and checks their health.
// if a device does not have an address or does not support the health check command, add it with status "not supported".
// and it's why it was not checked.
// Returns a slice of HealthStatus for each device in the room.
func GetDeviceHealth(ctx context.Context, roomID string) ([]HealthStatus, error) {
	devices, err := couchdb.GetDevicesByRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}

	results := make([]HealthStatus, 0, len(devices))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, dev := range devices {
		if len(dev.Address) == 0 ||
			strings.HasPrefix(dev.Address, "0.0.0.") ||
			!dev.HasCommand(healthCheckCmd) {
			continue
		}
		wg.Add(1)
		go func(d model.Device) {
			defer wg.Done()
			status := probe(ctx, d)
			mu.Lock()
			results = append(results, status)
			mu.Unlock()
		}(dev)
	}

	wg.Wait()
	return results, nil
}

// probe checks the health of a device by sending a GET request to its {Address}/health.
func probe(ctx context.Context, device model.Device) HealthStatus {
	hs := HealthStatus{DeviceID: device.ID}

	address, err := device.BuildCommandURL(healthCheckCmd)
	if err != nil {
		hs.Status = "error"
		hs.Error = fmt.Sprintf("unable to build command URL: %s", err.Error())
		return hs
	}
	// fill in the address
	address = strings.Replace(address, ":address", device.Address, 1)
<<<<<<< HEAD
	req, err := http.NewRequestWithContext(ctx, "GET", address, nil)
=======
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
>>>>>>> 8f6d957 (adding context to health check requests with timeouts)
	if err != nil {
		hs.Status = "error"
		hs.Error = fmt.Sprintf("unable to create request: %s", err.Error())
		return hs
	}
	req.Header.Set("User-Agent", "Device Monitoring Health Check")
<<<<<<< HEAD
	resp, err := healthHTTPClient.Do(req)
=======
	resp, err := deviceHealthHTTPClient.Do(req)
>>>>>>> 8f6d957 (adding context to health check requests with timeouts)
	if err != nil {
		hs.Status = "error"
		hs.Error = fmt.Sprintf("unable to check health: %s", err.Error())
		return hs
	}
	defer resp.Body.Close()

<<<<<<< HEAD
	bytes, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
=======
	bytes, err := io.ReadAll(io.LimitReader(resp.Body, maxHealthBody))
>>>>>>> 8f6d957 (adding context to health check requests with timeouts)
	if err != nil {
		hs.Status = "error"
		hs.Error = fmt.Sprintf("unable to read response: %s", err.Error())
		return hs
	}
	if resp.StatusCode != http.StatusOK {
		hs.Status = "error"
		hs.Error = fmt.Sprintf("health check failed. response: %s", bytes)
		return hs
	}
	hs.Status = healthyStatus
	// if the response is not empty, we can log it
	if len(bytes) > 0 {
		slog.Info("Health check response", slog.String("device_id", device.ID), slog.String("response", string(bytes)))
	}
	return hs
}

func GetRoomHealth(ctx context.Context, roomID string) ([]HealthStatus, error) {
	results, err := GetDeviceHealth(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device health: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no devices found in room %s", roomID)
	}

	return results, nil
}
