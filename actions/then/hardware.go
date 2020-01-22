package then

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/device-monitoring/actions/hardwareinfo"
	"github.com/byuoitav/device-monitoring/localsystem"
	"github.com/byuoitav/device-monitoring/messenger"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
)

func hardwareInfo(ctx context.Context, with []byte, log *zap.SugaredLogger) *nerr.E {
	systemID, err := localsystem.SystemID()
	if err != nil {
		return err.Addf("unable to get hardware info")
	}

	deviceInfo := events.GenerateBasicDeviceInfo(systemID)

	info, err := hardwareinfo.PiInfo()
	if err != nil {
		return err.Addf("unable to get hardware info")
	}

	// build base event
	event := events.Event{
		GeneratingSystem: systemID,
		Timestamp:        time.Now(),
		EventTags: []string{
			events.HardwareInfo,
		},
		TargetDevice: deviceInfo,
		AffectedRoom: deviceInfo.BasicRoomInfo,
		Key:          "hardware-info",
		Data:         info,
	}

	// send info dump
	messenger.Get().SendEvent(event)
	event.Data = nil

	if usage, ok := info.CPU["usage"].(map[string]float64); ok {
		if avg, ok := usage["avg"]; ok {
			tmp := event
			tmp.AddToTags(events.DetailState)
			tmp.Key = "cpu-usage-percent"
			tmp.Value = fmt.Sprintf("%v", avg)
			messenger.Get().SendEvent(tmp)
		}
	}

	// send info about cpu load averages
	if loadAvg1min, ok := info.CPU["avg1min"].(map[string]float64); ok {
		tmp := event
		tmp.AddToTags(events.DetailState)
		tmp.Key = "cpu-load-average-1-min"
		tmp.Value = fmt.Sprintf("%v", loadAvg1min)
		messenger.Get().SendEvent(tmp)
	}

	if loadAvg5min, ok := info.CPU["avg5min"].(map[string]float64); ok {
		tmp := event
		tmp.AddToTags(events.DetailState)
		tmp.Key = "cpu-load-average-5-min"
		tmp.Value = fmt.Sprintf("%v", loadAvg5min)
		messenger.Get().SendEvent(tmp)
	}

	// send info about memory usage
	if vMem, ok := info.Memory["virtual"].(*mem.VirtualMemoryStat); ok {
		tmp := event
		tmp.AddToTags(events.DetailState)
		tmp.Key = "v-mem-used-percent"
		tmp.Value = fmt.Sprintf("%v", vMem.UsedPercent)
		messenger.Get().SendEvent(tmp)
	}

	// send info about swap usage
	if sMem, ok := info.Memory["swap"].(*mem.SwapMemoryStat); ok {
		event.Key = "s-mem-used-percent"
		event.Value = fmt.Sprintf("%v", sMem.UsedPercent)
		messenger.Get().SendEvent(event)
	}

	// send info about chip temp
	if temps, ok := info.Host["temperature"].(map[string]float64); ok {
		for chip, temp := range temps {
			tmp := event
			tmp.AddToTags(events.DetailState)
			tmp.Key = fmt.Sprintf("%s-temp", chip)
			tmp.Value = fmt.Sprintf("%v", temp)
			messenger.Get().SendEvent(tmp)
		}
	}

	// send info about # of writes
	if counters, ok := info.Disk["io-counters"]; ok {
		if disks, ok := counters.(map[string]disk.IOCountersStat); ok {
			for disk, stats := range disks {
				tmp := event
				tmp.AddToTags(events.DetailState)
				tmp.Key = fmt.Sprintf("writes-to-%s", disk)
				tmp.Value = fmt.Sprintf("%v", stats.WriteCount)
				messenger.Get().SendEvent(tmp)
			}
		}
	}

	// send info about total disk usage
	if usage, ok := info.Disk["usage"].(*disk.UsageStat); ok {
		tmp := event
		tmp.AddToTags(events.DetailState)
		tmp.Key = "disk-used-percent"
		tmp.Value = fmt.Sprintf("%v", usage.UsedPercent)
		messenger.Get().SendEvent(tmp)
	}

	// send info about avg # of processes in uninterruptible sleep
	if avg, ok := info.Procs["avg-procs-u-sleep"]; ok {
		tmp := event
		tmp.AddToTags(events.DetailState)
		tmp.Key = "avg-procs-u-sleep"
		tmp.Value = fmt.Sprintf("%v", avg)
		messenger.Get().SendEvent(tmp)
	}

	return nil
}

func deviceHardwareInfo(ctx context.Context, with []byte, log *zap.SugaredLogger) *nerr.E {
	systemID, err := localsystem.SystemID()
	if err != nil {
		return err.Addf("unable to get device hardware info")
	}

	// timeout if this takes longer than 30 seconds
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	info, err := hardwareinfo.RoomDevicesInfo(ctx)
	if err != nil {
		return err.Addf("unable to get device hardware info")
	}

	// key: deviceID, value: structs.HardwareInfo
	for k, v := range info {
		// build base event
		deviceInfo := events.GenerateBasicDeviceInfo(k)
		event := events.Event{
			GeneratingSystem: systemID,
			Timestamp:        time.Now(),
			EventTags:        []string{events.HardwareInfo},
			TargetDevice:     deviceInfo,
			AffectedRoom:     deviceInfo.BasicRoomInfo,
			Key:              "hardware-info",
			Data:             v,
		}

		// send info dump for this device
		messenger.Get().SendEvent(event)
		event.Data = nil

		// push up hostname
		if len(v.Hostname) > 0 {
			tmp := event
			tmp.AddToTags(events.DetailState)
			tmp.Key = "hostname"
			tmp.Value = v.Hostname
			messenger.Get().SendEvent(tmp)
		}

		if len(v.ModelName) > 0 {
			tmp := event
			tmp.AddToTags(events.DetailState)
			tmp.Key = "hardware-version"
			tmp.Value = v.ModelName
			messenger.Get().SendEvent(tmp)
		}

		if len(v.SerialNumber) > 0 {
			tmp := event
			tmp.AddToTags(events.DetailState)
			tmp.Key = "serial-number"
			tmp.Value = v.SerialNumber
			messenger.Get().SendEvent(tmp)
		}

		if len(v.FirmwareVersion) > 0 {
			tmp := event
			tmp.AddToTags(events.DetailState)
			tmp.Key = "software-version"
			// TODO what kind of interface{}...?
			tmp.Value = fmt.Sprintf("%v", v.FirmwareVersion)
			messenger.Get().SendEvent(tmp)
		}

		if len(v.FilterStatus) > 0 {
			tmp := event
			tmp.AddToTags(events.DetailState)
			tmp.Key = "filter-status"
			tmp.Value = v.FilterStatus
			messenger.Get().SendEvent(tmp)
		}

		if len(v.WarningStatus) > 0 {
			tmp := event
			tmp.AddToTags(events.DetailState)
			tmp.Key = "warning-status"

			str := ""

			for i := range v.WarningStatus {
				str += v.WarningStatus[i]
			}

			tmp.Value = str
			messenger.Get().SendEvent(tmp)
		}

		if len(v.ErrorStatus) > 0 {
			tmp := event
			tmp.AddToTags(events.DetailState)
			tmp.Key = "error-status"
			str := ""

			for i := range v.ErrorStatus {
				str += v.WarningStatus[i]
			}

			tmp.Value = str
			messenger.Get().SendEvent(tmp)
		}

		if len(v.PowerStatus) > 0 {
			event.Key = "power-status"
			event.Value = v.PowerStatus
			messenger.Get().SendEvent(event)
		}

		if v.TimerInfo != nil {
			event.Key = "timer-v"

			// TODO what kind of interface{}?
			event.Value = fmt.Sprintf("%v", v.TimerInfo)
			messenger.Get().SendEvent(event)
		}

		if len(v.NetworkInfo.IPAddress) > 0 {
			event.Key = "ip-address"
			event.Value = v.NetworkInfo.IPAddress
			messenger.Get().SendEvent(event)
		}

		if len(v.NetworkInfo.MACAddress) > 0 {
			event.Key = "mac-address"
			event.Value = v.NetworkInfo.MACAddress
			messenger.Get().SendEvent(event)
		}

		if len(v.NetworkInfo.Gateway) > 0 {
			event.Key = "default-gateway"
			event.Value = v.NetworkInfo.Gateway
			messenger.Get().SendEvent(event)
		}

		if len(v.NetworkInfo.DNS) > 0 {
			event.Key = "dns-addresses"
			builder := strings.Builder{}

			for i := range v.NetworkInfo.DNS {
				builder.WriteString(v.NetworkInfo.DNS[i])

				if i != len(v.NetworkInfo.DNS)-1 {
					builder.WriteString(", ")
				}
			}

			event.Value = builder.String()
			messenger.Get().SendEvent(event)
		}
	}

	return nil
}

type tempLimits struct {
	CriticalThreshold float64 `json:"critical-threshold"`
	WarningThreshold  float64 `json:"warning-threshold"`
}

func liveTemperatureCheck1(ctx context.Context, with []byte, log *zap.SugaredLogger) *nerr.E {
	systemID, err := localsystem.SystemID()
	if err != nil {
		return err.Addf("unable to get hardware info")
	}

	deviceInfo := events.GenerateBasicDeviceInfo(systemID)

	var info hardwareinfo.HardwareInfo

	info.Host, err = localsystem.HostInfo()
	if err != nil {
		return err.Addf("failed to get hardware info")
	}

	var limits tempLimits
	er := json.Unmarshal(with, &limits)
	if er != nil {
		return nerr.Translate(er)
	}

	// build base event
	event := events.Event{
		GeneratingSystem: systemID,
		Timestamp:        time.Now(),
		EventTags: []string{
			events.HardwareInfo,
		},
		TargetDevice: deviceInfo,
		AffectedRoom: deviceInfo.BasicRoomInfo,
	}

	// send info about chip temp
	if temps, ok := info.Host["temperature"].(map[string]float64); ok {
		for chip, temp := range temps {
			if temp > limits.WarningThreshold {
				if temp > limits.CriticalThreshold {
					event.AddToTags(events.DetailState)
					event.Key = "temp-critical"
					event.Value = fmt.Sprintf("%v", temp)
					event.Data = chip
					messenger.Get().SendEvent(event)
				} else {
					event.AddToTags(events.DetailState)
					event.Key = "temp-warning"
					event.Value = fmt.Sprintf("%v", temp)
					event.Data = chip
					messenger.Get().SendEvent(event)
				}
			}
		}
	}

	return nil
}

func liveTemperatureCheck(ctx context.Context, with []byte, log *zap.SugaredLogger) *nerr.E {
	//checks every 5 seconds to see if the temp is high
	ticker1 := time.NewTicker(5 * time.Second)
	//gives us the temp every 5 minutes
	ticker2 := time.NewTicker(5 * time.Minute)

	systemID, err := localsystem.SystemID()
	if err != nil {
		return err.Addf("unable to get hardware info")
	}

	deviceInfo := events.GenerateBasicDeviceInfo(systemID)
	previousTemps := make(map[string]float64)

	var info hardwareinfo.HardwareInfo

	var limits tempLimits
	er := json.Unmarshal(with, &limits)
	if er != nil {
		return nerr.Translate(er)
	}

	event := events.Event{
		GeneratingSystem: systemID,
		EventTags: []string{
			events.HardwareInfo,
		},
		TargetDevice: deviceInfo,
		AffectedRoom: deviceInfo.BasicRoomInfo,
	}

	for {
		select {
		case t1 := <-ticker1.C:
			//check to see if the temp is over the warning or critical threshold

			info.Host, err = localsystem.HostInfo()
			if err != nil {
				return err.Addf("failed to get hardware info")
			}

			if temps, ok := info.Host["temperature"].(map[string]float64); ok {
				for chip, temp := range temps {
					oldTemp := previousTemps[chip]
					if temp > limits.WarningThreshold {
						if temp > limits.CriticalThreshold {
							if oldTemp < limits.CriticalThreshold {
								event.AddToTags(events.DetailState)
								event.Key = "temp-critical"
								event.Value = fmt.Sprintf("%v", temp)
								event.Data = chip
								event.Timestamp = t1
								messenger.Get().SendEvent(event)
							}
						} else {
							if oldTemp > limits.CriticalThreshold || oldTemp < limits.WarningThreshold {
								event.AddToTags(events.DetailState)
								event.Key = "temp-warning"
								event.Value = fmt.Sprintf("%v", temp)
								event.Data = chip
								event.Timestamp = t1
								messenger.Get().SendEvent(event)
							}
						}
					}
					previousTemps[chip] = temp
				}
			}

		case t2 := <-ticker2.C:
			//give us the 5 minute check

			info.Host, err = localsystem.HostInfo()
			if err != nil {
				return err.Addf("failed to get hardware info")
			}

			if temps, ok := info.Host["temperature"].(map[string]float64); ok {
				for chip, temp := range temps {
					if temp > limits.WarningThreshold {
						if temp > limits.CriticalThreshold {
							event.AddToTags(events.DetailState)
							event.Key = "temp-critical"
							event.Value = fmt.Sprintf("%v", temp)
							event.Data = chip
							event.Timestamp = t2
							messenger.Get().SendEvent(event)
						} else if temp > limits.WarningThreshold {
							event.AddToTags(events.DetailState)
							event.Key = "temp-warning"
							event.Value = fmt.Sprintf("%v", temp)
							event.Data = chip
							event.Timestamp = t2
							messenger.Get().SendEvent(event)
						} else {
							event.AddToTags(events.DetailState)
							event.Key = "temp-normal"
							event.Value = fmt.Sprintf("%v", temp)
							event.Data = chip
							event.Timestamp = t2
							messenger.Get().SendEvent(event)
						}
					}
				}
			}
		}
	}
}
