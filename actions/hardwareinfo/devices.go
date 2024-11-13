package hardwareinfo

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/device-monitoring/localsystem"
)

const (
	hardwareInfoCommandID = "HardwareInfo"
)

// RoomDevicesInfo .
func RoomDevicesInfo(ctx context.Context) (map[string]structs.HardwareInfo, *nerr.E) {
	info := make(map[string]structs.HardwareInfo)

	roomID, err := localsystem.RoomID()
	if err != nil {
		return info, err.Addf("failed to get hardware info about devices in room")
	}

	log.L.Infof("Getting hardware info about devices in room")

	devices, gerr := db.GetDB().GetDevicesByRoom(roomID)
	if gerr != nil {
		return info, nerr.Translate(gerr).Addf("failed to get hardware info about devices in %s", roomID)
	}

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	for i := range devices {
		// skip the pi's
		if devices[i].Type.ID == "Pi3" ||
			devices[i].Address == "0.0.0.0" ||
			len(devices[i].Address) == 0 ||
			!devices[i].HasCommand(hardwareInfoCommandID) {
			continue
		}

		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			i := getHardwareInfo(ctx, devices[idx])

			mu.Lock()
			info[devices[idx].ID] = i
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	return info, nil
}

func getHardwareInfo(ctx context.Context, device structs.Device) structs.HardwareInfo {
	var info structs.HardwareInfo

	address, err := device.BuildCommandURL(hardwareInfoCommandID)
	if err != nil {
		log.L.Warnf("unable to get hardware info for %s: %s", device.ID, err.Error())
		return info
	}

	address = strings.Replace(address, ":address", device.Address, 1)

	req, gerr := http.NewRequest("GET", address, nil)
	if gerr != nil {
		log.L.Warnf("unable to get hardware info for %s: %s", device.ID, gerr)
		return info
	}

	req = req.WithContext(ctx)
	c := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, gerr := c.Do(req)
	if gerr != nil {
		log.L.Warnf("unable to get hardware info for %s: %s", device.ID, gerr)
		return info
	}
	defer resp.Body.Close()

	bytes, gerr := io.ReadAll(resp.Body)
	if gerr != nil {
		log.L.Warnf("unable to get hardware info for %s: %s", device.ID, gerr)
		return info
	}

	gerr = json.Unmarshal(bytes, &info)
	if gerr != nil {
		log.L.Warnf("unable to get hardware info for %s: %s", device.ID, gerr)
		return info
	}

	return info
}
