package hardwareinfo

import (
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/device-monitoring/localsystem"
)

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

// PiInfo .
func PiInfo() (HardwareInfo, *nerr.E) {
	log.L.Infof("Getting pi hardware info")
	var info HardwareInfo
	var err *nerr.E

	info.CPU, err = localsystem.CPUInfo()
	if err != nil {
		return info, err.Addf("failed to get hardware info")
	}

	info.Memory, err = localsystem.MemoryInfo()
	if err != nil {
		return info, err.Addf("failed to get hardware info")
	}

	info.Host, err = localsystem.HostInfo()
	if err != nil {
		return info, err.Addf("failed to get hardware info")
	}

	info.Disk, err = localsystem.DiskInfo()
	if err != nil {
		return info, err.Addf("failed to get hardware info")
	}

	info.Network, err = localsystem.NetworkInfo()
	if err != nil {
		return info, err.Addf("failed to get hardware info")
	}

	info.Docker, err = localsystem.DockerInfo()
	if err != nil {
		return info, err.Addf("failed to get hardware info")
	}

	info.Procs, err = localsystem.ProcsInfo()
	if err != nil {
		return info, err.Addf("failed to get hardware info")
	}

	return info, nil
}
