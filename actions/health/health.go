package health

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/device-monitoring/couchdb"
	"github.com/byuoitav/device-monitoring/model"
)

const (
	HealthEndpoint = "/health"
	healthyStatus  = "healthy"
	healthCheckCmd = "HealthCheck"
)

// HealthStatus is the JSON‑serializable result for one device.
type HealthStatus struct {
	DeviceID string `json:"device_id"`
	Status   string `json:"status"` // "healthy" or "error" or "not supported"
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
		if dev.Address == "" || !dev.HasCommand(healthCheckCmd) {
			results = append(results, HealthStatus{
				DeviceID: dev.ID,
				Status:   "not supported",
				Error:    "device has no address or does not support health check command",
			})
			continue
		}
		wg.Add(1)
		go func(d model.Device) {
			defer wg.Done()
			status := probe(d)
			mu.Lock()
			results = append(results, status)
			mu.Unlock()
		}(dev)
	}

	wg.Wait()
	return results, nil
}

// probe checks the health of a device by sending a GET request to its {Address}/health.
func probe(device model.Device) HealthStatus {
	hs := HealthStatus{DeviceID: device.ID}

	url := strings.TrimRight(device.Address, "/") + HealthEndpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		hs.Status = "error"
		hs.Error = err.Error()
		return hs
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		hs.Status = "error"
		hs.Error = err.Error()
		return hs
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		hs.Status = "error"
		hs.Error = err.Error()
		return hs
	}
	if resp.StatusCode/100 != 2 {
		hs.Status = "error"
		hs.Error = fmt.Sprintf("status %d: %s", resp.StatusCode, string(body))
		return hs
	}

	hs.Status = healthyStatus
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
