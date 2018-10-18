package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/byuoitav/authmiddleware"
	"github.com/byuoitav/common"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/device-monitoring/handlers"
	"github.com/byuoitav/device-monitoring/jobs"
	"github.com/labstack/echo"
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

	secure := router.Group("", echo.WrapMiddleware(authmiddleware.Authenticate))

	// device info endpoints
	secure.GET("/device", handlers.GetDeviceInfo)
	secure.GET("/device/hostname", handlers.GetHostname)
	secure.GET("/device/id", handlers.GetDeviceID)
	secure.GET("/device/ip", handlers.GetIPAddress)
	secure.GET("/device/network", handlers.IsConnectedToInternet)
	secure.GET("/device/mstatus", handlers.GetMStatusInfo)

	secure.GET("/room", handlers.GetRoom)
	secure.GET("/room/state", handlers.RoomState)
	secure.GET("/room/ping", handlers.PingStatus)

	// action endpoints
	secure.PUT("/device/reboot", handlers.RebootPi)

	// dashboard
	// TODO redirect from /dash
	secure.Static("/dashboard", "ui/dashboard")

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}
	router.StartServer(&server)
}
