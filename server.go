package main

import (
	"context"
	"net/http"

	"github.com/byuoitav/common"
	"github.com/byuoitav/device-monitoring/actions"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	_ "github.com/byuoitav/device-monitoring/actions/then"
)

func main() {
	go actions.ActionManager().Start(context.TODO())

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

	/*
		// device info endpoints
		router.GET("/device", handlers.GetDeviceInfo)
		router.GET("/device/hostname", handlers.GetHostname)
		router.GET("/device/id", handlers.GetDeviceID)
		router.GET("/device/ip", handlers.GetIPAddress)
		router.GET("/device/network", handlers.IsConnectedToInternet)
		router.GET("/device/status", handlers.GetStatusInfo)
		router.GET("/device/dhcp", handlers.GetDHCPState)
		router.GET("/device/screenshot", handlers.GetScreenshot)
		router.GET("/device/hardwareinfo", handlers.GetMyHardwareInfo)
		router.GET("/device/runners", handlers.GetRunnerInfo)

		// room info endpoints
		router.GET("/room", handlers.GetRoom)
		router.GET("/room/state", handlers.RoomState)
		router.GET("/room/activesignal", handlers.ActiveSignal)
		router.GET("/room/hardwareinfo", handlers.DeviceHardwareInfo)
		router.GET("/room/ping", handlers.PingStatus)
		router.GET("/room/viainfo", handlers.ViaInfo)

		// divider endpoints
		router.GET("/divider/state", handlers.GetDividerState)
		router.GET("/divider/preset/:hostname", handlers.PresetForHostname)

		// action endpoints
		router.PUT("/device/reboot", handlers.RebootPi)
		router.PUT("/device/dhcp/:state", handlers.SetDHCPState)
		router.POST("/event", handlers.SendEvent)

		// test mode endpoints
		// router.GET("/maintenance", handlers.IsInMaintMode)
		// router.PUT("/maintenance", handlers.ToggleMaintMode)

		// provisioning endpoints
		router.GET("/provisioning/ws", socket.UpgradeToWebsocket(provisioning.SocketManager()))
		router.GET("/provisioning/id", handlers.GetProvisioningID)
	*/

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}
	router.StartServer(&server)
}
