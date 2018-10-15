package handlers

import (
	"fmt"
	"net/http"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/device-monitoring/jobs"
	"github.com/byuoitav/device-monitoring/jobs/ask"
	"github.com/byuoitav/device-monitoring/pi"
	"github.com/labstack/echo"
)

// GetDeviceInfo combines lots of device information into a response.
func GetDeviceInfo(context echo.Context) error {
	data := make(map[string]interface{})

	internet := pi.IsConnectedToInternet()
	data["internet-connectivity"] = internet

	hostname, err := pi.Hostname()
	if err != nil {
		data["error"] = err
		return context.JSON(http.StatusInternalServerError, data)
	}
	data["hostname"] = hostname

	ip, err := pi.IPAddress()
	if err != nil {
		data["error"] = err
		return context.JSON(http.StatusInternalServerError, data)
	}
	data["ip"] = ip

	id, err := pi.DeviceID()
	if err != nil {
		data["error"] = err
		return context.JSON(http.StatusInternalServerError, data)
	}
	data["id"] = id

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

// RoomState returns the room state, but also pulses it around the room
func RoomState(context echo.Context) error {
	// pulse the room state
	job := &ask.StateUpdateJob{}
	state := jobs.RunJob(job, nil)

	switch v := state.(type) {
	case error:
		return context.String(http.StatusInternalServerError, v.Error())
	case *nerr.E:
		return context.String(http.StatusInternalServerError, v.Error())
	case nerr.E:
		return context.String(http.StatusInternalServerError, v.Error())
	case base.PublicRoom:
		return context.JSON(http.StatusOK, v)
	case *base.PublicRoom:
		return context.JSON(http.StatusOK, v)
	default:
		return context.String(http.StatusInternalServerError, fmt.Sprintf("something went wrong: %v", v))
	}
}
