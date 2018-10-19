package handlers

import (
	"fmt"
	"net/http"

	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/status"
	"github.com/byuoitav/device-monitoring/jobs"
	"github.com/byuoitav/device-monitoring/jobs/ask"
	"github.com/byuoitav/device-monitoring/pi"
	"github.com/labstack/echo"
)

// GetDeviceInfo combines lots of device information into a response.
func GetDeviceInfo(context echo.Context) error {
	data := make(map[string]interface{})

	// internet status
	internet := pi.IsConnectedToInternet()
	data["internet-connectivity"] = internet

	// device hostname
	hostname, err := pi.Hostname()
	if err != nil {
		data["error"] = err
		return context.JSON(http.StatusInternalServerError, data)
	}
	data["hostname"] = hostname

	// device id
	id, err := pi.DeviceID()
	if err != nil {
		data["error"] = err
		return context.JSON(http.StatusInternalServerError, data)
	}
	data["id"] = id

	// device ip address
	ip, err := pi.IPAddress()
	if err != nil {
		data["error"] = err
		return context.JSON(http.StatusInternalServerError, data)
	}
	data["ip"] = ip

	// status
	job := &ask.StatusJob{}
	jobContext := jobs.GetJobContext("status")

	s := jobs.RunJob(job, jobContext)

	switch v := s.(type) {
	case error:
		data["error"] = v
		return context.JSON(http.StatusInternalServerError, data)
	case *nerr.E:
		data["error"] = v.String()
		return context.JSON(http.StatusInternalServerError, data)
	case []status.Status:
		data["status"] = v
	default:
		data["error"] = fmt.Sprintf("unable to get status: unexpected type from job: %v", v)
		return context.JSON(http.StatusInternalServerError, data)
	}

	return context.JSON(http.StatusOK, data)
}

// GetHostname returns the hostname of the device we are on
func GetHostname(context echo.Context) error {
	hostname, err := pi.Hostname()
	if err != nil {
		return context.String(http.StatusInternalServerError, err.Error())
	}
	return context.String(http.StatusOK, hostname)
}

// GetDeviceID returns the hostname of the device we are on
func GetDeviceID(context echo.Context) error {
	id, err := pi.DeviceID()
	if err != nil {
		return context.String(http.StatusInternalServerError, err.Error())
	}
	return context.String(http.StatusOK, id)
}

// GetIPAddress returns the ip address of the device we are on
func GetIPAddress(context echo.Context) error {
	ip, err := pi.IPAddress()
	if err != nil {
		return context.String(http.StatusInternalServerError, err.Error())
	}
	return context.String(http.StatusOK, ip.String())
}

// IsConnectedToInternet returns a bool of true/false
func IsConnectedToInternet(context echo.Context) error {
	status := pi.IsConnectedToInternet()
	return context.String(http.StatusOK, fmt.Sprintf("%v", status))
}

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
