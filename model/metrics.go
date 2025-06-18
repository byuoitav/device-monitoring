package model

type HardwareInfo struct {
	Hostname              string           `json:"hostname,omitempty"`
	ModelName             string           `json:"model_name,omitempty"`
	SerialNumber          string           `json:"serial_number,omitempty"`
	BuildDate             string           `json:"build_date,omitempty"`
	FirmwareVersion       string           `json:"firmware_version,omitempty"`
	ProtocolVersion       string           `json:"protocol_version,omitempty"`
	NetworkInfo           NetworkInfo      `json:"network_information,omitempty"`
	FilterStatus          string           `json:"filter_status,omitempty"`
	WarningSatus          []string         `json:"warning_status,omitempty"`
	ErrorStatus           []string         `json:"error_status,omitempty"`
	PowerStatus           string           `json:"power_status,omitempty"`
	PowerSavingModeStatus string           `json:"power_saving_mode_status,omitempty"`
	TimerInfo             []map[string]int `json:"timer_information,omitempty"`
	Temperature           string           `json:"temperature,omitempty"`
}

type NetworkInfo struct {
	IPAddress  string `json:"ip_address,omitempty"`
	MACAddress string `json:"mac_address,omitempty"`
	Gateway    string `json:"gateway,omitempty"`
	DNS        string `json:"dns,omitempty"`
}
