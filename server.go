package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/byuoitav/common"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/device-monitoring/handlers"
	"github.com/byuoitav/device-monitoring/jobs"
)

func main() {
	// for some reason, after sending icmp packets, you can't kill the service without this
	// catch sigterm and exit
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.L.Infof("Captured sigterm")
		os.Exit(1)
	}()

	// start jobs
	go jobs.StartJobScheduler()

	// server
	port := ":10000"
	router := common.NewRouter()

	// secure := router.Group("", echo.WrapMiddleware(authmiddleware.Authenticate))

	// device info endpoints
	router.GET("/device", handlers.GetDeviceInfo)
	router.GET("/device/hostname", handlers.GetHostname)
	router.GET("/device/id", handlers.GetDeviceID)
	router.GET("/device/ip", handlers.GetIPAddress)
	router.GET("/device/network", handlers.IsConnectedToInternet)
	router.GET("/device/status", handlers.GetStatusInfo)
	router.GET("/device/dhcp", handlers.GetDHCPState)

	router.GET("/room", handlers.GetRoom)
	router.GET("/room/state", handlers.RoomState)
	router.GET("/room/ping", handlers.PingStatus)

	// action endpoints
	router.PUT("/device/reboot", handlers.RebootPi)
	router.PUT("/device/dhcp/:state", handlers.SetDHCPState)

	// dashboard
	// TODO redirect from /dash
	router.Static("/dashboard", "dashboard")

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}
	router.StartServer(&server)
}
