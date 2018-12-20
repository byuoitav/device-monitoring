package ask

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/device-monitoring/localsystem"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/docker"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

const (
	temperatureRootPath = "/sys/class/thermal"
	uSleepCheckInterval = 3 * time.Second
)

// HardwareInfoJob gets hardware information about the device and pushes events up about it
type HardwareInfoJob struct{}

// HardwareInfo is the struct of hardware information that is returned by this job.
type HardwareInfo struct {
	Host    map[string]interface{} `json:"host,omitempty"`
	Memory  map[string]interface{} `json:"memory,omitempty"`
	CPU     map[string]interface{} `json:"cpu,omitempty"`
	Disk    map[string]interface{} `json:"disk,omitempty"`
	Network map[string]interface{} `json:"network,omitempty"`
	Docker  map[string]interface{} `json:"docker,omitempty"`
	Procs   map[string]interface{} `json:"procs,omitempty"`
}

var (
	avgProcsInUSleep float64
)

func init() {
	// constantly measure the number of processes that are in sleep and get an average
	avgProcsInUSleep = 0

	ticker := time.NewTicker(uSleepCheckInterval)

	go func() {
		defer ticker.Stop()

		for range ticker.C {
			procs, err := process.Processes()
			if err != nil {
				log.L.Warnf("failed to get running processes: %s", err)
				continue
			}

			count := 0

			for _, p := range procs {
				status, err := p.Status()
				if err != nil {
					continue
				}

				if status == "D" {
					count++
				}
			}

			avgProcsInUSleep = (avgProcsInUSleep + float64(count)) / 2
			avgProcsInUSleep = round(avgProcsInUSleep, .05)
		}
	}()
}

// Run runs the job
func (j *HardwareInfoJob) Run(ctx interface{}, eventWrite chan events.Event) interface{} {
	ret := HardwareInfo{}
	err := &nerr.E{}

	log.L.Infof("Getting Hardware Info")

	ret.CPU, err = getCPUInfo()
	if err != nil {
		return err.Addf("failed to get hardware info")
	}

	ret.Memory, err = getMemoryInfo()
	if err != nil {
		return err.Addf("failed to get hardware info")
	}

	ret.Host, err = getHostInfo()
	if err != nil {
		return nerr.Translate(err).Addf("failed to get hardware info")
	}

	ret.Disk, err = getDiskInfo()
	if err != nil {
		return nerr.Translate(err).Addf("failed to get hardware info")
	}

	ret.Network, err = getNetworkInfo()
	if err != nil {
		return nerr.Translate(err).Addf("failed to get hardware info")
	}

	ret.Docker, err = getDockerInfo()
	if err != nil {
		return nerr.Translate(err).Addf("failed to get hardware info")
	}

	ret.Procs, err = getProcsInfo()
	if err != nil {
		return nerr.Translate(err).Addf("failed to get hardware info")
	}

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

func getCPUInfo() (map[string]interface{}, *nerr.E) {
	info := make(map[string]interface{})

	// get hardware info about cpu
	cpuState, err := cpu.Info()
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get cpu info")
	}

	info["hardware"] = cpuState

	// get percent usage information per cpu
	usage := make(map[string]float64)
	info["usage"] = usage

	percentages, err := cpu.Percent(0, true)
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get cpu info")
	}

	for i := range percentages {
		usage[fmt.Sprintf("cpu%d", i)] = round(percentages[i], .01)
	}

	// get average usage
	avgPercent, err := cpu.Percent(0, false)
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get cpu info")
	}

	if len(avgPercent) == 1 {
		usage["avg"] = round(avgPercent[0], .01)
	}

	return info, nil
}

func getMemoryInfo() (map[string]interface{}, *nerr.E) {
	info := make(map[string]interface{})

	vMem, err := mem.VirtualMemory()
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get memory info")
	}

	vMem.UsedPercent = round(vMem.UsedPercent, .01)
	info["virtual"] = vMem

	sMem, err := mem.SwapMemory()
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get memory info")
	}

	sMem.UsedPercent = round(sMem.UsedPercent, .01)
	info["swap"] = sMem

	return info, nil
}

func getHostInfo() (map[string]interface{}, *nerr.E) {
	info := make(map[string]interface{})

	stat, err := host.Info()
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get host info")
	}

	info["os"] = stat

	users, err := host.Users()
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get host info")
	}

	info["users"] = users

	temps := make(map[string]float64)
	count := make(map[string]int)
	info["temperature"] = temps

	filepath.Walk(temperatureRootPath, func(path string, info os.FileInfo, err error) error {
		if info.Mode()&os.ModeSymlink == os.ModeSymlink && strings.Contains(path, "thermal_") {
			// get type
			ttype, err := ioutil.ReadFile(path + "/type")
			if err != nil {
				return err
			}

			// get temperature
			ttemp, err := ioutil.ReadFile(path + "/temp")
			if err != nil {
				return err
			}

			stype := strings.TrimSpace(string(ttype))
			dtemp, err := strconv.ParseFloat(strings.TrimSpace(string(ttemp)), 64)

			temps[fmt.Sprintf("%s%d", stype, count[stype])] = dtemp / 1000
			count[stype]++
		}

		if info.IsDir() && path != temperatureRootPath {
			return filepath.SkipDir
		}

		return nil
	})

	return info, nil
}

func getDiskInfo() (map[string]interface{}, *nerr.E) {
	info := make(map[string]interface{})

	usage, err := disk.Usage("/")
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get host info")
	}

	usage.UsedPercent = round(usage.UsedPercent, .01)
	info["usage"] = usage

	ioCounters, err := disk.IOCounters("sda", "mmcblk0")
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get host info")
	}

	info["io-counters"] = ioCounters

	return info, nil
}

func getNetworkInfo() (map[string]interface{}, *nerr.E) {
	info := make(map[string]interface{})

	interfaces, err := net.Interfaces()
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get host info")
	}

	info["interfaces"] = interfaces

	return info, nil
}

func getDockerInfo() (map[string]interface{}, *nerr.E) {
	info := make(map[string]interface{})

	stats, err := docker.GetDockerStat()
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get host info")
	}

	info["stats"] = stats

	return info, nil
}

func getProcsInfo() (map[string]interface{}, *nerr.E) {
	info := make(map[string]interface{})

	procs, err := process.Processes()
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get processes info")
	}

	bad := []string{}

	for _, p := range procs {
		status, err := p.Status()
		if err != nil {
			continue
		}

		if status == "D" {
			name, err := p.Name()
			if err != nil {
				name = fmt.Sprintf("unable to get name: %s", name)
			}

			bad = append(bad, name)
		}
	}

	info["cur-procs-u-sleep"] = bad
	info["avg-procs-u-sleep"] = avgProcsInUSleep

	return info, nil
}

func round(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}
