package ask

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/device-monitoring/localsystem"
)

const (
	// TODO change this command name to match whatever command we need
	hardwareInfoCommandID = "HardwareInfo"
)

// DeviceHardwareJob gets hardware information from devices in the room and pushes it up
type DeviceHardwareJob struct{}

// Run runs the job
func (j *DeviceHardwareJob) Run(ctx interface{}, eventWrite chan events.Event) interface{} {
	log.L.Infof("Getting hardware info for devices in room")

	// get list of devices from database
	roomID, err := localsystem.RoomID()
	if err != nil {
		return err.Addf("failed to get hardware info about devices")
	}

	devices, gerr := db.GetDB().GetDevicesByRoom(roomID)
	if gerr != nil {
		return nerr.Translate(gerr).Addf("failed to get hardware info about devices in %s", roomID)
	}

	wg := sync.WaitGroup{}
	hardwareInfo := make(map[string]structs.HardwareInfo)
	mu := sync.Mutex{}

	for i := range devices {
		// skip the pi's
		if devices[i].Type.ID == "Pi3" {
			continue
		}
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			info := getHardwareInfo(&devices[idx])
			if info != nil {
				sendHardwareInfo(devices[idx].ID, info, eventWrite)

				mu.Lock()
				hardwareInfo[devices[idx].ID] = *info
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	return hardwareInfo
}

func getHardwareInfo(device *structs.Device) *structs.HardwareInfo {
	if device == nil {
		log.L.Errorf("device to get hardware info from cannot be null")
		return nil
	}

	address := device.GetCommandByID(hardwareInfoCommandID).BuildCommandAddress()
	if len(address) == 0 {
		log.L.Infof("%s doesn't have a %s command, so I can't get any hardware info about it", device.ID, hardwareInfoCommandID)
		return nil
	}

	log.L.Infof("Getting hardware info for %s", device.ID)

	address = strings.Replace(address, ":address", device.Address, 1)

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	log.L.Debugf("Sending GET request to: %s", address)

	// get hardware info about device
	resp, err := client.Get(address)
	if err != nil {
		log.L.Warnf("failed to get hardware info for %s: %s", device.ID, err)
		return nil
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.L.Warnf("failed to get hardware info for %s: %s", device.ID, err)
		return nil
	}

	ret := &structs.HardwareInfo{}

	err = json.Unmarshal(bytes, ret)
	if err != nil {
		log.L.Warnf("failed to get hardware info for %s: %s", device.ID, err)
		return nil
	}

	return ret
}

func sendHardwareInfo(deviceID string, info *structs.HardwareInfo, eventWrite chan events.Event) {
	// push up events about device
	targetDevice := events.GenerateBasicDeviceInfo(deviceID)
	event := events.Event{
		GeneratingSystem: localsystem.MustSystemID(),
		Timestamp:        time.Now(),
		EventTags: []string{
			events.HardwareInfo,
		},
		TargetDevice: targetDevice,
		AffectedRoom: events.GenerateBasicRoomInfo(targetDevice.RoomID),
		Key:          "hardware-info",
		Data:         info,
	}
	eventWrite <- event // dump up all the hardware info
	event.Data = nil

	// push up hostname
	if len(info.Hostname) > 0 {
		tmp := event
		tmp.AddToTags(events.DetailState)
		tmp.Key = "hostname"
		tmp.Value = info.Hostname
		eventWrite <- tmp
	}

	if len(info.ModelName) > 0 {
		tmp := event
		tmp.AddToTags(events.DetailState)
		tmp.Key = "model-name"
		tmp.Value = info.ModelName
		eventWrite <- tmp
	}

	if len(info.SerialNumber) > 0 {
		tmp := event
		tmp.AddToTags(events.DetailState)
		tmp.Key = "serial-number"
		tmp.Value = info.SerialNumber
		eventWrite <- tmp
	}

	if len(info.FirmwareVersion) > 0 {
		tmp := event
		tmp.AddToTags(events.DetailState)
		tmp.Key = "firmware-version"
		// TODO what kind of interface{}...?
		tmp.Value = fmt.Sprintf("%v", info.FirmwareVersion)
		eventWrite <- tmp
	}

	if len(info.FilterStatus) > 0 {
		tmp := event
		tmp.AddToTags(events.DetailState)
		tmp.Key = "filter-status"
		tmp.Value = info.FilterStatus
		eventWrite <- tmp
	}

	if len(info.WarningStatus) > 0 {
		tmp := event
		tmp.AddToTags(events.DetailState)
		tmp.Key = "warning-status"

		str := ""

		for i := range info.WarningStatus {
			str += info.WarningStatus[i]
		}

		tmp.Value = str
		eventWrite <- tmp
	}

	if len(info.ErrorStatus) > 0 {
		tmp := event
		tmp.AddToTags(events.DetailState)
		tmp.Key = "error-status"
		str := ""

		for i := range info.ErrorStatus {
			str += info.WarningStatus[i]
		}

		tmp.Value = str
		eventWrite <- tmp
	}

	if len(info.PowerStatus) > 0 {
		event.Key = "power-status"
		event.Value = info.PowerStatus
		eventWrite <- event
	}

	if info.TimerInfo != nil {
		event.Key = "timer-info"

		// TODO what kind of interface{}?
		event.Value = fmt.Sprintf("%v", info.TimerInfo)
		eventWrite <- event
	}

	if len(info.NetworkInfo.IPAddress) > 0 {
		event.Key = "ip-address"
		event.Value = info.NetworkInfo.IPAddress
		eventWrite <- event
	}

	if len(info.NetworkInfo.MACAddress) > 0 {
		event.Key = "mac-address"
		event.Value = info.NetworkInfo.MACAddress
		eventWrite <- event
	}

	if len(info.NetworkInfo.Gateway) > 0 {
		event.Key = "default-gateway"
		event.Value = info.NetworkInfo.Gateway
		eventWrite <- event
	}

	if len(info.NetworkInfo.DNS) > 0 {
		event.Key = "dns-addresses"
		builder := strings.Builder{}

		for i := range info.NetworkInfo.DNS {
			builder.WriteString(info.NetworkInfo.DNS[i])

			if i != len(info.NetworkInfo.DNS)-1 {
				builder.WriteString(", ")
			}
		}

		event.Value = builder.String()
		eventWrite <- event
	}
}
