package main

import (
	"net/http"

	"github.com/byuoitav/common"
	"github.com/byuoitav/device-monitoring/handlers"
	"github.com/byuoitav/device-monitoring/jobs"
	"github.com/byuoitav/device-monitoring/provisioning"
	"github.com/byuoitav/device-monitoring/socket"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	// start jobs
	go jobs.StartJobScheduler()

	// server
	port := ":10000"
	router := common.NewRouter()

	// remove this eventually
	// redirect from /dash to /dashboard
	router.GET("/dash", func(context echo.Context) error {
		return context.Redirect(http.StatusMovedPermanently, "/dashboard")
	})

	// static webpages
	router.Group("/dashboard", middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "dashboard",
		Index:  "index.html",
		HTML5:  true,
		Browse: true,
	}))

	// device info endpoints
	router.GET("/device", handlers.GetDeviceInfo)
	router.GET("/device/hostname", handlers.GetHostname)
	router.GET("/device/id", handlers.GetDeviceID)
	router.GET("/device/ip", handlers.GetIPAddress)
	router.GET("/device/network", handlers.IsConnectedToInternet)
	router.GET("/device/status", handlers.GetStatusInfo)
	router.GET("/device/dhcp", handlers.GetDHCPState)
	router.GET("/device/hardwareinfo", handlers.GetMyHardwareInfo)

	// router.GET("/hardwareinfo/:id", handlers.GetHardwareInfo)

	router.GET("/room", handlers.GetRoom)
	router.GET("/room/state", handlers.RoomState)
	router.GET("/room/ping", handlers.PingStatus)

	// action endpoints
	router.PUT("/device/reboot", handlers.RebootPi)
	router.PUT("/device/dhcp/:state", handlers.SetDHCPState)

	// provisioning endpoints
	router.GET("/provisioning/ws", socket.UpgradeToWebsocket(provisioning.SocketManager()))
	router.GET("/provisioning/id", handlers.GetProvisioningID)

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}
	router.StartServer(&server)
}
