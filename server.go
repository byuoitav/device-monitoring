package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/byuoitav/authmiddleware"
	"github.com/byuoitav/device-monitoring-microservice/device"
	"github.com/byuoitav/device-monitoring-microservice/statemonitoring"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {

	//get building and room info
	hostname := os.Getenv("PI_HOSTNAME")
	building := strings.Split(hostname, "-")[0]
	room := strings.Split(hostname, "-")[1]

	statemonitoring.StartPublisher()

	statemonitoring.StartMonitoring(time.Second*300, "localhost:8000", building, room)

	//get addresses from database
	devices, err := device.GetAddresses(building, room)
	if err != nil {
		log.Printf("Error getting devices from database: %s", err.Error())
	}

	//figure out how often to ping devices and start process in new goroutine
	pingInterval := os.Getenv("DEVICE_PING_INTERVAL")
	interval, err := strconv.Atoi(pingInterval)
	if err != nil {

		log.Printf("Error reading check interval. Terminating...")
		os.Exit(1)

	} else {

		go func() {

			for {

				device.PingAddresses(building, room, devices)
				time.Sleep(time.Duration(interval) * time.Second)

			}

		}()
	}

	port := ":10000"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	secure := router.Group("", echo.WrapMiddleware(authmiddleware.Authenticate))

	secure.GET("/health", Health)

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	router.StartServer(&server)

}

func Health(context echo.Context) error {
	return context.JSON(http.StatusOK, "The fleet has moved out of lightspeed and we're preparing to - augh!")
}
