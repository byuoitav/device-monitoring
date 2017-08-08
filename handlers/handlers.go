package handlers

import (
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo"
)

type HealthResponse struct {
	Version string `json:"version"`
	Status  int    `json:"statuscode"`
}

func GetHostname(context echo.Context) error {
	pihn := os.Getenv("PI_HOSTNAME")
	if len(pihn) == 0 {
		return context.JSON(http.StatusInternalServerError, "PI Hostname not set")
	}

	return context.JSON(http.StatusOK, pihn)
}

func GetIP(context echo.Context) error {
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

	log.Printf("My IP address is %v", ip.String())
	return context.JSON(http.StatusOK, ip.String())
}

func GetNetworkConnectedStatus(context echo.Context) error {
	_, err := net.Dial("tcp", "google.com:80")
	if err != nil {
		return context.JSON(http.StatusInternalServerError, err.Error())
	}

	return context.JSON(http.StatusOK, true)
}
