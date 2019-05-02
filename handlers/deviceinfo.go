package handlers

import (
	"fmt"
	"net/http"

	"github.com/byuoitav/device-monitoring/actions/hardwareinfo"
	"github.com/byuoitav/device-monitoring/actions/screenshot"
	"github.com/byuoitav/device-monitoring/localsystem"
	"github.com/labstack/echo"
)

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
	// TODO this endpoint
	return nil
}

/*
// GetStatusInfo returns the default status info
func GetStatusInfo(context echo.Context) error {
	job := &ask.StatusJob{}
	jobContext := jobs.GetJobContext("status")

	s := jobs.RunJob(job, jobContext)

	switch v := s.(type) {
	case error:
		return context.String(http.StatusInternalServerError, v.Error())
	case *nerr.E:
		return context.String(http.StatusInternalServerError, v.Error())
	case []status.Status:
		return context.JSON(http.StatusOK, v)
	default:
		return context.String(http.StatusInternalServerError, fmt.Sprintf("unexpected type from job: %v", v))
	}
}
*/
