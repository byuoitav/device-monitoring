package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/device-monitoring/actions/hardwareinfo"
	"github.com/byuoitav/device-monitoring/actions/health"
	"github.com/byuoitav/device-monitoring/actions/screenshot"
	"github.com/byuoitav/device-monitoring/localsystem"
	"github.com/labstack/echo"
)

// DeviceInfo .
type DeviceInfo struct {
	Hostname             string `json:"hostname,omitempty"`
	ID                   string `json:"id,omitempty"`
	IP                   string `json:"ip,omitempty"`
	InternetConnectivity bool   `json:"internet-connectivity"`

	DHCPInfo struct {
		Enabled    bool `json:"enabled"`
		Toggleable bool `json:"toggleable"`
	} `json:"dhcp"`
}

// GetDeviceInfo .
func GetDeviceInfo(ectx echo.Context) error {
	var info DeviceInfo
	var err *nerr.E

	info.Hostname, err = localsystem.Hostname()
	if err != nil {
		return ectx.String(http.StatusInternalServerError, err.Error())
	}

	info.ID, err = localsystem.SystemID()
	if err != nil {
		return ectx.String(http.StatusInternalServerError, err.Error())
	}

	ip, err := localsystem.IPAddress()
	if err != nil {
		return ectx.String(http.StatusInternalServerError, err.Error())
	}

	info.IP = ip.String()
	info.InternetConnectivity = localsystem.IsConnectedToInternet()

	info.DHCPInfo.Enabled, err = localsystem.UsingDHCP()
	if err != nil {
		return ectx.String(http.StatusInternalServerError, err.Error())
	}

	err = localsystem.CanToggleDHCP()
	if err != nil {
		info.DHCPInfo.Toggleable = false
	} else {
		info.DHCPInfo.Toggleable = true
	}

	return ectx.JSON(http.StatusOK, info)
}

// GetHostname returns the hostname of the device we are on
func GetHostname(ectx echo.Context) error {
	hostname, err := localsystem.Hostname()
	if err != nil {
		return ectx.String(http.StatusInternalServerError, err.Error())
	}

	return ectx.String(http.StatusOK, hostname)
}

// GetDeviceID returns the hostname of the device we are on
func GetDeviceID(ectx echo.Context) error {
	id, err := localsystem.SystemID()
	if err != nil {
		return ectx.String(http.StatusInternalServerError, err.Error())
	}

	return ectx.String(http.StatusOK, id)
}

// GetIPAddress returns the ip address of the device we are on
func GetIPAddress(ectx echo.Context) error {
	ip, err := localsystem.IPAddress()
	if err != nil {
		return ectx.String(http.StatusInternalServerError, err.Error())
	}

	return ectx.String(http.StatusOK, ip.String())
}

// IsConnectedToInternet returns a bool of true/false
func IsConnectedToInternet(ectx echo.Context) error {
	status := localsystem.IsConnectedToInternet()
	return ectx.String(http.StatusOK, fmt.Sprintf("%v", status))
}

// GetDHCPState returns whether or not dhcp is enabled and if it can be toggled or not
func GetDHCPState(ectx echo.Context) error {
	ret := make(map[string]interface{})

	usingDHCP, err := localsystem.UsingDHCP()
	if err != nil {
		ret["error"] = err.String()
		return ectx.JSON(http.StatusInternalServerError, ret)
	}
	ret["enabled"] = usingDHCP

	if err = localsystem.CanToggleDHCP(); err != nil {
		ret["error"] = err.String()
		return ectx.JSON(http.StatusInternalServerError, ret)
	}
	ret["toggleable"] = true

	return ectx.JSON(http.StatusOK, ret)
}

// GetScreenshot a screenshot of the device's screen
func GetScreenshot(ectx echo.Context) error {
	bytes, err := screenshot.Take(ectx.Request().Context())
	if err != nil {
		return ectx.String(http.StatusInternalServerError, err.Error())
	}

	return ectx.Blob(http.StatusOK, "image/jpeg", bytes)
}

// HardwareInfo returns hardware info about this device
func HardwareInfo(ectx echo.Context) error {
	info, err := hardwareinfo.PiInfo()
	if err != nil {
		return ectx.String(http.StatusInternalServerError, err.Error())
	}

	return ectx.JSON(http.StatusOK, info)
}

// GetServiceHealth returns the health of services on this device
func GetServiceHealth(ectx echo.Context) error {
	var configs []health.ServiceCheckConfig
	err := ectx.Bind(&configs)
	if err != nil {
		return ectx.String(http.StatusInternalServerError, err.Error())
	}

	// timeout if this takes longer than 15 seconds
	ctx, cancel := context.WithTimeout(ectx.Request().Context(), 15*time.Second)
	defer cancel()

	resps := health.CheckServices(ctx, configs)
	return ectx.JSON(http.StatusOK, resps)
}
