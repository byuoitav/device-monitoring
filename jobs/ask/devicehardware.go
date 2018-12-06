package ask

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
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
	// get list of devices from database
	roomID, err := localsystem.RoomID()
	if err != nil {
		return err.Addf("failed to get hardware info about devices")
	}

	devices, gerr := db.GetDB().GetDevicesByRoom(roomID)
	if gerr != nil {
		return nerr.Translate(gerr).Addf("failed to get hardware info about devices in %s", roomID)
	}

	for i := range devices {
		// skip the pi's
		if devices[i].Type.ID == "Pi3" {
			continue
		}

		for j := range devices[i].Type.Commands {
			if devices[i].Type.Commands[j].ID == hardwareInfoCommandID {
				go sendHardwareInfoForDevice(devices[i], devices[i].Type.Commands[j], eventWrite)
			}
		}
	}

	return nil
}

func sendHardwareInfoForDevice(device structs.Device, command structs.Command, eventWrite chan events.Event) {
	if command.ID != hardwareInfoCommandID {
		log.L.Warnf("unable to send hardware info for %s because the wrong command (%s) was passed in.", device.ID, command.ID)
		return
	}

	// build the endpoint to hit
	url := fmt.Sprintf("%s%s", command.Microservice.Address, command.Endpoint.Path)

	// replace parameters
	if strings.Contains(url, ":address") {
		url = strings.Replace(url, ":address", device.Address, 1)
	}

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	// get hardware info about device
	resp, err := client.Get(url)
	if err != nil {
		log.L.Warnf("failed to get hardware info for %s: %s", device.ID, err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.L.Warnf("failed to get hardware info for %s: %s", device.ID, err)
		return
	}

	info := structs.HardwareInfo{}
	err = json.Unmarshal(body, &info)
	if err != nil {
		log.L.Warnf("failed to get hardware info for %s: %s", device.ID, err)
		return
	}

	// push up events about device
	targetDevice := events.GenerateBasicDeviceInfo(device.ID)
	event := events.Event{
		GeneratingSystem: localsystem.MustSystemID(),
		Timestamp:        time.Now(),
		EventTags:        []string{
			// TODO add tags
		},
		TargetDevice: targetDevice,
		AffectedRoom: events.GenerateBasicRoomInfo(targetDevice.RoomID),
		Key:          "hardware-info",
		Value:        "",
		User:         "",
		Data:         info,
	}
	eventWrite <- event // dump up all the hardware info
	event.Data = nil
	event.Key = ""

	// push up hostname
	if len(info.Hostname) > 0 {
		event.Key = "hostname"
		event.Value = info.Hostname
		eventWrite <- event
	}

	if len(info.ModelName) > 0 {
		event.Key = "model-name"
		event.Value = info.ModelName
		eventWrite <- event
	}

	if len(info.SerialNumber) > 0 {
		event.Key = "serial-number"
		event.Value = info.SerialNumber
		eventWrite <- event
	}

	if info.FirmwareVersion != nil {
		event.Key = "firmware-version"
		// TODO what kind of interface{}...?
		event.Value = fmt.Sprintf("%v", info.FirmwareVersion)
		eventWrite <- event
	}

	if len(info.FilterStatus) > 0 {
		event.Key = "filter-status"
		event.Value = info.FilterStatus
		eventWrite <- event
	}

	if len(info.WarningStatus) > 0 {
		event.Key = "warning-status"
		// TODO why is this one an []
		// event.Value = info.WarningStatus
		eventWrite <- event
	}

	if len(info.ErrorStatus) > 0 {
		event.Key = "error-status"
		// TODO why is this one an []
		// event.Value = info.ErrorStatus
		eventWrite <- event
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
