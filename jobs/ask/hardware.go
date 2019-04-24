package ask

/*
const (
	temperatureRootPath = "/sys/class/thermal"
	uSleepCheckInterval = 3 * time.Second
	uSleepResetInterval = 5 * time.Minute
)

// HardwareInfoJob gets hardware information about the device and pushes events up about it
type HardwareInfoJob struct{}

// Run runs the job
func (j *HardwareInfoJob) Run(ctx interface{}, eventWrite chan events.Event) interface{} {
	ret := HardwareInfo{}
	err := &nerr.E{}

	log.L.Infof("Getting Hardware Info")

	systemID, err := localsystem.SystemID()
	if err != nil {
		log.L.Warnf("SYSTEM_ID not set, so I wont send any hardware events.")
		return ret
	}

	roomID, err := localsystem.RoomID()
	if err != nil {
		log.L.Warnf("SYSTEM_ID not set, so I wont send any hardware events.")
		return ret
	}

	// send event with all the information
	event := events.Event{
		GeneratingSystem: systemID,
		Timestamp:        time.Now(),
		EventTags: []string{
			events.HardwareInfo,
		},
		TargetDevice: events.GenerateBasicDeviceInfo(systemID),
		AffectedRoom: events.GenerateBasicRoomInfo(roomID),
		Key:          "hardware-info",
		Data:         ret,
	}
	eventWrite <- event
	event.Data = nil

	// send info about cpu usage
	if usage, ok := ret.CPU["usage"].(map[string]float64); ok {
		if avg, ok := usage["avg"]; ok {
			tmp := event
			tmp.AddToTags(events.DetailState)
			tmp.Key = "cpu-usage-percent"
			tmp.Value = fmt.Sprintf("%v", avg)
			eventWrite <- tmp
		}
	}

	// send info about memory usage
	if vMem, ok := ret.Memory["virtual"].(*mem.VirtualMemoryStat); ok {
		tmp := event
		tmp.AddToTags(events.DetailState)
		tmp.Key = "v-mem-used-percent"
		tmp.Value = fmt.Sprintf("%v", vMem.UsedPercent)
		eventWrite <- tmp
	}

	// send info about swap usage
	if sMem, ok := ret.Memory["swap"].(*mem.SwapMemoryStat); ok {
		event.Key = "s-mem-used-percent"
		event.Value = fmt.Sprintf("%v", sMem.UsedPercent)
		eventWrite <- event
	}

	// send info about chip temp
	if temps, ok := ret.Host["temperature"].(map[string]float64); ok {
		for chip, temp := range temps {
			tmp := event
			tmp.AddToTags(events.DetailState)
			tmp.Key = fmt.Sprintf("%s-temp", chip)
			tmp.Value = fmt.Sprintf("%v", temp)
			eventWrite <- tmp
		}
	}

	// send info about # of writes
	if counters, ok := ret.Disk["io-counters"]; ok {
		if disks, ok := counters.(map[string]disk.IOCountersStat); ok {
			for disk, stats := range disks {
				tmp := event
				tmp.AddToTags(events.DetailState)
				tmp.Key = fmt.Sprintf("writes-to-%s", disk)
				tmp.Value = fmt.Sprintf("%v", stats.WriteCount)
				eventWrite <- tmp
			}
		}
	}

	// send info about total disk usage
	if usage, ok := ret.Disk["usage"].(*disk.UsageStat); ok {
		tmp := event
		tmp.AddToTags(events.DetailState)
		tmp.Key = "disk-used-percent"
		tmp.Value = fmt.Sprintf("%v", usage.UsedPercent)
		eventWrite <- tmp
	}

	// send info about avg # of processes in uninterruptible sleep
	if avg, ok := ret.Procs["avg-procs-u-sleep"]; ok {
		tmp := event
		tmp.AddToTags(events.DetailState)
		tmp.Key = "avg-procs-u-sleep"
		tmp.Value = fmt.Sprintf("%v", avg)
		eventWrite <- tmp
	}

	return ret
}
*/
