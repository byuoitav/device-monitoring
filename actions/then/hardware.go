package then

import (
	"context"
	"fmt"
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
