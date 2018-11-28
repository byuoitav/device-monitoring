package ask

import (
	"fmt"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/v2/events"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

const (
	temperatureRootPath = "/sys/class/thermal"
)

// HardwareInfoJob gets hardware information about the device and pushes events up about it
type HardwareInfoJob struct{}

// HardwareInfo is the struct of hardware information that is returned by this job.
type HardwareInfo struct {
	Host   map[string]interface{} `json:"host,omitempty"`
	Memory map[string]interface{} `json:"memory,omitemtpy"`
	CPU    map[string]interface{} `json:"cpu,omitempty"`
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
		usage[fmt.Sprintf("cpu%d", i)] = percentages[i]
	}

	// get average usage
	avgPercent, err := cpu.Percent(0, false)
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get cpu info")
	}

	if len(avgPercent) == 1 {
		usage["cpu"] = avgPercent[0]
	}

	return info, nil
}

func getMemoryInfo() (map[string]interface{}, *nerr.E) {
	info := make(map[string]interface{})

	vMem, err := mem.VirtualMemory()
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get memory info")
	}

	info["virtual"] = vMem

	sMem, err := mem.SwapMemory()
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get memory info")
	}

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

	// get temperature info
	temps, err := host.SensorsTemperatures()
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get host info")
	}

	info["temperature"] = temps

	users, err := host.Users()
	if err != nil {
		return info, nerr.Translate(err).Addf("failed to get host info")
	}

	info["users"] = users

	return info, nil
}
