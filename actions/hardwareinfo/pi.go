package hardwareinfo

import (
	"fmt"
	"log/slog"

	"github.com/byuoitav/device-monitoring/localsystem"
)

// HardwareInfo .
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
func PiInfo() (HardwareInfo, error) {
	slog.Info("Getting pi hardware info")
	var info HardwareInfo
	var err error

	info.CPU, err = localsystem.CPUInfo()
	if err != nil {
		return info, fmt.Errorf("failed to get hardware info: %w", err)
	}

	info.Memory, err = localsystem.MemoryInfo()
	if err != nil {
		return info, fmt.Errorf("failed to get hardware info: %w", err)
	}

	info.Host, err = localsystem.HostInfo()
	if err != nil {
		return info, fmt.Errorf("failed to get hardware info: %w", err)
	}

	info.Disk, err = localsystem.DiskInfo()
	if err != nil {
		return info, fmt.Errorf("failed to get hardware info: %w", err)
	}

	info.Network, err = localsystem.NetworkInfo()
	if err != nil {
		return info, fmt.Errorf("failed to get hardware info: %w", err)
	}

	info.Docker, err = localsystem.DockerInfo()
	if err != nil {
		return info, fmt.Errorf("failed to get hardware info: %w", err)
	}

	info.Procs, err = localsystem.ProcsInfo()
	if err != nil {
		return info, fmt.Errorf("failed to get hardware info: %w", err)
	}

	return info, nil
}
