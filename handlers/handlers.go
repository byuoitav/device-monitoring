package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/device-monitoring-microservice/pi"
	"github.com/labstack/echo"
)

// GetDeviceInfo combines lots of device information into a response.
func GetDeviceInfo(context echo.Context) error {
	return nil
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

// RebootPi reboots the pi
func RebootPi(context echo.Context) error {
	go func() {
		for i := 5; i > 0; i-- {
			log.L.Infof("REBOOTING PI IN %v SECONDS", i)
			time.Sleep(1 * time.Second)
		}

		err := pi.Reboot()
		if err != nil {
			log.L.Errorf("failed to reboot pi: %v", err.Error())
		}
	}()

	return context.Blob(http.StatusOK, "text/plain", []byte("Rebooting in 5 seconds..."))
}
