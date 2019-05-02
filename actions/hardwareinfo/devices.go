package hardwareinfo

import (
	"context"
	"encoding/json"
	"io/ioutil"
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
			len(devices[i].GetCommandByID(hardwareInfoCommandID).ID) == 0 {
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

	address := device.GetCommandByID(hardwareInfoCommandID).BuildCommandAddress()
	if len(address) == 0 {
		log.L.Warnf("unable to build command address for %s command on %s", hardwareInfoCommandID, device.ID)
		return info
	}

	log.L.Infof("Getting hardware info for %s", device.ID)
	address = strings.Replace(address, ":address", device.Address, 1)

	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		log.L.Warnf("unable to get hardware info for %s: %s", device.ID, err)
		return info
	}

	req = req.WithContext(ctx)
	c := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := c.Do(req)
	if err != nil {
		log.L.Warnf("unable to get hardware info for %s: %s", device.ID, err)
		return info
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.L.Warnf("unable to get hardware info for %s: %s", device.ID, err)
		return info
	}

	err = json.Unmarshal(bytes, &info)
	if err != nil {
		log.L.Warnf("unable to get hardware info for %s: %s", device.ID, err)
		return info
	}

	return info
}
