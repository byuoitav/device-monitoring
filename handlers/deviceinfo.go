package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"log/slog"

	"github.com/byuoitav/device-monitoring/actions/hardwareinfo"
	"github.com/byuoitav/device-monitoring/actions/health"
	"github.com/byuoitav/device-monitoring/actions/screenshot"
	"github.com/byuoitav/device-monitoring/localsystem"
	"github.com/gin-gonic/gin"
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
func GetDeviceInfo(c *gin.Context) {
	var info DeviceInfo
	var err error

	if info.Hostname, err = localsystem.Hostname(); err != nil {
		slog.Error("hostname lookup failed", slog.Any("error", err))
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if info.ID, err = localsystem.SystemID(); err != nil {
		slog.Error("system ID lookup failed", slog.Any("error", err))
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	ip, err := localsystem.IPAddress()
	if err != nil {
		slog.Error("ip address lookup failed", slog.Any("error", err))
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	info.IP = ip.String()
	info.InternetConnectivity = localsystem.IsConnectedToInternet()

	if info.DHCPInfo.Enabled, err = localsystem.UsingDHCP(); err != nil {
		slog.Error("failed to check DHCP state", slog.Any("error", err))
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if err = localsystem.CanToggleDHCP(); err != nil {
		info.DHCPInfo.Toggleable = false
	} else {
		info.DHCPInfo.Toggleable = true
	}

	c.JSON(http.StatusOK, info)
}

// GetHostname returns the hostname of the device we are on
func GetHostname(c *gin.Context) {
	hostname, err := localsystem.Hostname()
	if err != nil {
		slog.Error("hostname lookup failed", slog.Any("error", err))
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, hostname)
}

// GetSystemID returns the system ID of the device we are on
func GetDeviceID(c *gin.Context) {
	id, err := localsystem.SystemID()
	if err != nil {
		slog.Error("system ID lookup failed", slog.Any("error", err))
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, id)
}

// GetIPAddress returns the ip address of the device we are on
func GetIPAddress(c *gin.Context) {
	ipAddr, err := localsystem.IPAddress()
	if err != nil {
		slog.Error("ip address lookup failed", slog.Any("error", err))
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, ipAddr.String())
}

// IsConnectedToInternet returns a bool of true/false
func IsConnectedToInternet(c *gin.Context) {
	status := localsystem.IsConnectedToInternet()
	c.String(http.StatusOK, fmt.Sprintf("%v", status))
}

// GetDHCPState returns whether or not dhcp is enabled and if it can be toggled or not
func GetDHCPState(c *gin.Context) {
	ret := make(map[string]interface{})
	enabled, err := localsystem.UsingDHCP()
	if err != nil {
		slog.Error("UsingDHCP failed", slog.Any("error", err))
		ret["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, ret)
		return
	}
	ret["enabled"] = enabled

	if err := localsystem.CanToggleDHCP(); err != nil {
		slog.Info("Cannot toggle DHCP", slog.Any("error", err))
		ret["toggleable"] = false
	} else {
		ret["toggleable"] = true
	}

	c.JSON(http.StatusOK, ret)
}

// GetScreenshot a screenshot of the device's screen
func GetScreenshot(c *gin.Context) {
	imgBytes, err := screenshot.Take(c.Request.Context())
	if err != nil {
		slog.Error("screenshot failed", slog.Any("error", err))
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Data(http.StatusOK, "image/jpeg", imgBytes)
}

// HardwareInfo returns hardware info about this device
func HardwareInfo(c *gin.Context) {
	info, err := hardwareinfo.PiInfo()
	if err != nil {
		slog.Error("hardware info retrieval failed", slog.Any("error", err))
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, info)
}

// GetServiceHealth returns the health of services on this device
func GetServiceHealth(c *gin.Context) {
	var configs []health.ServiceCheckConfig
	if err := c.Bind(&configs); err != nil {
		slog.Error("failed to bind health config", slog.Any("error", err))
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	results := health.CheckServices(ctx, configs)
	c.JSON(http.StatusOK, results)
}
