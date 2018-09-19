package handlers

import (
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/labstack/echo"
)

// GetHostname returns the hostname of the device we are on
func GetHostname(context echo.Context) error {
	hostname, err := os.Hostname()
	if err != nil {
		return context.Blob(http.StatusInternalServerError, "text/plain", []byte(err.Error()))
	}

	return context.Blob(http.StatusOK, "text/plain", []byte(hostname))
}

// GetPiHostname returns the hostname of the device we are on
func GetPiHostname(context echo.Context) error {
	pihn := os.Getenv("PI_HOSTNAME")
	if len(pihn) == 0 {
		return context.Blob(http.StatusInternalServerError, "text/plain", []byte("PI_HOSTNAME not set."))
	}

	// TODO validate PI_HOSTNAME is in correct format

	return context.Blob(http.StatusOK, "text/plain", []byte(pihn))
}

// GetIPAddress returns the ip address of the device we are on
func GetIPAddress(context echo.Context) error {
	var ip net.IP
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return context.JSON(http.StatusInternalServerError, err.Error())
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && strings.Contains(address.String(), "/24") {
			ip, _, err = net.ParseCIDR(address.String())
			if err != nil {
				return context.JSON(http.StatusInternalServerError, err.Error())
			}
		}
	}

	if ip == nil {
		return context.JSON(http.StatusInternalServerError, "IP Address not found")
	}

	log.L.Infof("My IP address is %v", ip.String())
	return context.Blob(http.StatusOK, "text/plain", []byte(ip.String()))
}

func GetNetworkConnectedStatus(context echo.Context) error {
	_, err := net.Dial("tcp", "google.com:80")
	if err != nil {
		return context.JSON(http.StatusInternalServerError, err.Error())
	}

	return context.JSON(http.StatusOK, true)
}

func RebootPi(context echo.Context) error {
	defer color.Unset()
	color.Set(color.FgRed, color.Bold)
	log.Printf("\n\n\nRebooting Pi\n\n\n")

	http.Get("http://localhost:7010/reboot")
	return nil
}
