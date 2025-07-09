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

	"log/slog"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	dockerstat "github.com/shirou/gopsutil/docker"
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

// CPUInfo returns per-CPU and load-average stats.
func CPUInfo() (map[string]any, error) {
	info := make(map[string]any)

	cpuState, err := cpu.Info()
	if err != nil {
		slog.Error("failed to get CPU info", slog.Any("error", err))
		return info, fmt.Errorf("failed to get CPU info: %w", err)
	}
	info["hardware"] = cpuState

	percentages, err := cpu.Percent(0, true)
	if err != nil {
		slog.Error("failed to get per-CPU usage", slog.Any("error", err))
		return info, fmt.Errorf("failed to get CPU usage: %w", err)
	}
	usage := make(map[string]float64, len(percentages))
	for i, p := range percentages {
		usage[fmt.Sprintf("cpu%d", i)] = round(p, .01)
	}
	info["usage"] = usage

	avgPercent, err := cpu.Percent(0, false)
	if err != nil {
		slog.Error("failed to get average CPU usage", slog.Any("error", err))
		return info, fmt.Errorf("failed to get average CPU usage: %w", err)
	}
	if len(avgPercent) > 0 {
		usage["avg"] = round(avgPercent[0], .01)
	}

	loadAvg, err := load.Avg()
	if err != nil {
		slog.Error("failed to get load average", slog.Any("error", err))
		return info, fmt.Errorf("failed to get load average: %w", err)
	}
	info["avg1min"] = loadAvg.Load1
	info["avg5min"] = loadAvg.Load5

	return info, nil
}

// MemoryInfo returns virtual and swap memory statistics.
func MemoryInfo() (map[string]interface{}, error) {
	info := make(map[string]interface{})

	vMem, err := mem.VirtualMemory()
	if err != nil {
		slog.Error("failed to get virtual memory info", slog.Any("error", err))
		return info, fmt.Errorf("failed to get virtual memory info: %w", err)
	}
	vMem.UsedPercent = round(vMem.UsedPercent, .01)
	info["virtual"] = vMem

	sMem, err := mem.SwapMemory()
	if err != nil {
		slog.Error("failed to get swap memory info", slog.Any("error", err))
		return info, fmt.Errorf("failed to get swap memory info: %w", err)
	}
	sMem.UsedPercent = round(sMem.UsedPercent, .01)
	info["swap"] = sMem

	return info, nil
}

// HostInfo returns OS info, logged-in users, and thermal sensor readings.
func HostInfo() (map[string]interface{}, error) {
	info := make(map[string]interface{})

	stat, err := host.Info()
	if err != nil {
		slog.Error("failed to get host info", slog.Any("error", err))
		return info, fmt.Errorf("failed to get host info: %w", err)
	}
	info["os"] = stat

	users, err := host.Users()
	if err != nil {
		slog.Error("failed to get host users", slog.Any("error", err))
		return info, fmt.Errorf("failed to get host users: %w", err)
	}
	info["users"] = users

	temps := make(map[string]float64)
	count := make(map[string]int)
	info["temperature"] = temps

	if err := filepath.Walk(temperatureRootPath, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink && strings.Contains(path, "thermal_") {
			ttype, terr := os.ReadFile(path + "/type")
			if terr != nil {
				return terr
			}
			ttemp, terr := os.ReadFile(path + "/temp")
			if terr != nil {
				return terr
			}
			stype := strings.TrimSpace(string(ttype))
			dtemp, perr := strconv.ParseFloat(strings.TrimSpace(string(ttemp)), 64)
			if perr != nil {
				return perr
			}
			key := fmt.Sprintf("%s%d", stype, count[stype])
			temps[key] = dtemp / 1000
			count[stype]++
		}
		if fi.IsDir() && path != temperatureRootPath {
			return filepath.SkipDir
		}
		return nil
	}); err != nil {
		slog.Warn("error walking temperature sensors", slog.Any("error", err))
	}

	return info, nil
}

// DiskInfo returns usage and IO counters for key devices.
func DiskInfo() (map[string]interface{}, error) {
	info := make(map[string]interface{})

	usage, err := disk.Usage("/")
	if err != nil {
		slog.Error("failed to get disk usage", slog.Any("error", err))
		return info, fmt.Errorf("failed to get disk usage: %w", err)
	}
	usage.UsedPercent = round(usage.UsedPercent, .01)
	info["usage"] = usage

	ioCounters, err := disk.IOCounters("sda", "mmcblk0")
	if err != nil {
		slog.Error("failed to get disk IO counters", slog.Any("error", err))
		return info, fmt.Errorf("failed to get disk IO counters: %w", err)
	}
	info["io-counters"] = ioCounters

	return info, nil
}

// NetworkInfo returns the list of network interfaces.
func NetworkInfo() (map[string]interface{}, error) {
	info := make(map[string]interface{})

	ifaces, err := net.Interfaces()
	if err != nil {
		slog.Error("failed to get network interfaces", slog.Any("error", err))
		return info, fmt.Errorf("failed to get network interfaces: %w", err)
	}
	info["interfaces"] = ifaces

	return info, nil
}

// DockerInfo returns Docker stats and the count of running containers.
func DockerInfo() (map[string]interface{}, error) {
	info := make(map[string]interface{})

	stats, err := dockerstat.GetDockerStat()
	if err != nil {
		slog.Error("failed to get Docker stats", slog.Any("error", err))
		return info, fmt.Errorf("failed to get Docker stats: %w", err)
	}
	info["stats"] = stats

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		slog.Error("failed to create Docker client", slog.Any("error", err))
		return info, fmt.Errorf("failed to create Docker client: %w", err)
	}
	cli.NegotiateAPIVersion(context.Background())

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		slog.Error("failed to list Docker containers", slog.Any("error", err))
		return info, fmt.Errorf("failed to list Docker containers: %w", err)
	}
	info["docker-containers"] = len(containers)

	return info, nil
}

// ProcsInfo returns the names of processes in uninterruptible sleep,
// plus a running average of how many are in that state.
func ProcsInfo() (map[string]interface{}, error) {
	avgProcsInit.Do(startWatchingUSleep)
	info := make(map[string]interface{})

	procs, err := process.Processes()
	if err != nil {
		slog.Error("failed to list processes", slog.Any("error", err))
		return info, fmt.Errorf("failed to list processes: %w", err)
	}

	var bad []string
	for _, p := range procs {
		status, serr := p.Status()
		if serr != nil {
			continue
		}
		if status == "D" {
			if name, nerr := p.Name(); nerr == nil {
				bad = append(bad, name)
			} else {
				bad = append(bad, fmt.Sprintf("unknown(%v)", p.Pid))
			}
		}
	}

	info["cur-procs-u-sleep"] = bad
	info["avg-procs-u-sleep"] = avgProcsInUSleep
	return info, nil
}

// startWatchingUSleep continuously measures processes in uninterruptible sleep.
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
					slog.Warn("failed to list processes for u-sleep monitoring", slog.Any("error", err))
					continue
				}
				count := 0
				for _, p := range procs {
					if status, serr := p.Status(); serr == nil && status == "D" {
						count++
					}
				}
				avgProcsInUSleep = round((avgProcsInUSleep+float64(count))/2, .05)

			case <-resetTicker.C:
				avgProcsInUSleep = 0
			}
		}
	}()
}

// round to the nearest multiple of unit.
func round(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}
