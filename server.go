package main

import (
	"log"
	"net/http"

	"github.com/byuoitav/device-monitoring-microservice/device"
	"github.com/labstack/echo"
)

func main() {
	devices, err := device.GetAddresses("ITB", "1101")
	if err != nil {
		log.Printf("Houston, we have a problem.")
	}

	device.Bash(devices)

	//	device.ScanNetwork()
}

func Health(context echo.Context) error {
	return context.JSON(http.StatusOK, "The fleet has moved out of lightspeed and we're preparing to - augh!")
}
