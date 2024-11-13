package localsystem

import (
	"context"
	"fmt"
	"math"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/docker"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

const (
	temperatureRootPath = "/sys/class/thermal"
	uSleepCheckInterval = 3 * time.Second
	uSleepResetInterval = 5 * time.Minute
)

var (
	avgProcsInit     sync.Once
	avgProcsInUSleep float64
)

// CPUInfo .
func CPUInfo() (map[string]interface{}, *nerr.E) {
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

	// get load average metrics
	loadAvg, err := load.Avg()
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get load avg info")
	}

	info["avg1min"] = loadAvg.Load1
	info["avg5min"] = loadAvg.Load5

	return info, nil
}

// MemoryInfo .
func MemoryInfo() (map[string]interface{}, *nerr.E) {
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

// HostInfo .
func HostInfo() (map[string]interface{}, *nerr.E) {
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
			ttype, err := os.ReadFile(path + "/type")
			if err != nil {
				return err
			}

			// get temperature
			ttemp, err := os.ReadFile(path + "/temp")
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

// DiskInfo .
func DiskInfo() (map[string]interface{}, *nerr.E) {
	info := make(map[string]interface{})

	usage, err := disk.Usage("/")
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get disk info")
	}

	usage.UsedPercent = round(usage.UsedPercent, .01)
	info["usage"] = usage

	ioCounters, err := disk.IOCounters("sda", "mmcblk0")
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get disk info")
	}

	info["io-counters"] = ioCounters

	return info, nil
}

// NetworkInfo .
func NetworkInfo() (map[string]interface{}, *nerr.E) {
	info := make(map[string]interface{})

	interfaces, err := net.Interfaces()
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get network info")
	}

	info["interfaces"] = interfaces

	return info, nil
}

// DockerInfo .
func DockerInfo() (map[string]interface{}, *nerr.E) {
	info := make(map[string]interface{})

	stats, err := docker.GetDockerStat()
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get docker info")
	}

	info["stats"] = stats

	//add section getting the number of running docker containers
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get docker info")
	}

	containerList := types.ContainerListOptions{} // reset the container list

	cli.NegotiateAPIVersion(ctx)
	containers, err := cli.ContainerList(context.Background(), containerList)
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get docker info")
	}

	info["docker-containers"] = len(containers)

	return info, nil
}

// ProcsInfo .
func ProcsInfo() (map[string]interface{}, *nerr.E) {
	avgProcsInit.Do(startWatchingUSleep)
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

// constantly measure the number of processes that are in sleep and get an average
func startWatchingUSleep() {
	avgProcsInUSleep = 0

	checkTicker := time.NewTicker(uSleepCheckInterval)
	resetTicker := time.NewTicker(uSleepResetInterval)

	go func() {
		defer checkTicker.Stop()
		defer resetTicker.Stop()

		for {
			select {
			case <-checkTicker.C:
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
			case <-resetTicker.C:
				avgProcsInUSleep = 0
			}
		}
	}()
}

func round(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}
