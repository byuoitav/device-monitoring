package handlers

import (
	"fmt"
	"net/http"
	"sync"

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
		return context.Blob(http.StatusInternalServerError, "text/plain", []byte(err.Error()))
	}
	return context.Blob(http.StatusOK, "text/plain", []byte(hostname))
}

// GetDeviceID returns the hostname of the device we are on
func GetDeviceID(context echo.Context) error {
	id, err := pi.DeviceID()
	if err != nil {
		return context.Blob(http.StatusInternalServerError, "text/plain", []byte(err.Error()))
	}
	return context.Blob(http.StatusOK, "text/plain", []byte(id))
}

// GetIPAddress returns the ip address of the device we are on
func GetIPAddress(context echo.Context) error {
	ip, err := pi.IPAddress()
	if err != nil {
		return context.Blob(http.StatusInternalServerError, "text/plain", []byte(err.Error()))
	}
	return context.Blob(http.StatusOK, "text/plain", []byte(ip.String()))
}

// IsConnectedToInternet returns a bool of true/false
func IsConnectedToInternet(context echo.Context) error {
	status := pi.IsConnectedToInternet()
	return context.Blob(http.StatusOK, "text/plain", []byte(fmt.Sprintf("%v", status)))
}

// RoomState returns the room state, but also pulses it around the room
func RoomState(context echo.Context) error {
	wg := sync.WaitGroup{}

	// pulse the room state
	job := ask.StateUpdateJob{}
	job.Run(nil, nil)

	return context.JSON(http.StatusOK, "")
}
