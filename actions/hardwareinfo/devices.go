package hardwareinfo

import (
	"context"
	"encoding/json"
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
	hardwareInfoCommandID = "Hardware_Info"
)

// RoomDevicesInfo queries every non‑Pi device in the room for its hardware info.
// Returns a map of deviceID -> HardwareInfo, or an error if the room lookup or DB call fails.
func RoomDevicesInfo(ctx context.Context) (map[string]model.HardwareInfo, error) {
	// get the current room ID
	roomID, err := localsystem.RoomID()
	if err != nil {
		return nil, fmt.Errorf("failed to get room ID: %w", err)
	}

	slog.Info("Getting hardware info for devices in room", slog.String("room_id", roomID))

	// fetch devices
	devices, err := couchdb.GetDevicesByRoom(ctx, roomID)
	//devices, err := db.GetDB().GetDevicesByRoom(roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices in room %q: %w", roomID, err)
	}
	slog.Info("Devices fetched", slog.Int("count", len(devices)), slog.String("room_id", roomID))

	// concurrently collect each device’s hardware info
	infoMap := make(map[string]model.HardwareInfo, len(devices))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, dev := range devices {
		// skip Pis, zero‑address, or devices without the command
		if dev.Type.ID == "Pi3" ||
			len(dev.Address) == 0 ||
			dev.Address == "0.0.0.0" ||
			!dev.HasCommand(hardwareInfoCommandID) {
			slog.Info("Skipping device", slog.String("id", dev.ID), slog.String("type", dev.Type.ID), slog.String("address", dev.Address), slog.Bool("has_command", dev.HasCommand(hardwareInfoCommandID)))
			continue
		}

		wg.Add(1)
		go func(d model.Device) {
			defer wg.Done()
			hw := getHardwareInfo(ctx, d)

			mu.Lock()
			infoMap[d.ID] = hw
			mu.Unlock()
		}(dev)
	}

	wg.Wait()
	return infoMap, nil
}

// getHardwareInfo invokes the HardwareInfo command on a single device.
// Logs warnings on error and returns an empty HardwareInfo on failure.
func getHardwareInfo(ctx context.Context, device model.Device) model.HardwareInfo {
	var info model.HardwareInfo

	var err error
	// build command URL
	url, err := device.BuildCommandURL(hardwareInfoCommandID)
	if err != nil {
		slog.Warn("failed to build hardware info URL",
			slog.String("device", device.ID),
			slog.Any("error", err),
		)
		return info
	}
	url = strings.Replace(url, ":address", device.Address, 1)

	// execute the command
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		slog.Warn("failed to create hardware info request",
			slog.String("device", device.ID),
			slog.String("error", err.Error()),
		)
		return info
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		slog.Warn("hardware info request failed",
			slog.String("device", device.ID),
			slog.String("error", fmt.Sprintf("%v", err)),
		)
		return info
	}
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	// read and parse
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		slog.Warn("failed to read hardware info response",
			slog.String("device", device.ID),
			slog.Any("error", readErr),
		)
		return info
	}

	if err := json.Unmarshal(body, &info); err != nil {
		slog.Warn("failed to unmarshal hardware info JSON",
			slog.String("device", device.ID),
			slog.Any("error", err),
		)
		return info
	}

	return info
}
