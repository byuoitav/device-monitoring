package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/byuoitav/common"
	"github.com/byuoitav/device-monitoring/actions"
	"github.com/byuoitav/device-monitoring/handlers"
	"github.com/byuoitav/device-monitoring/messenger"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/pflag"

	_ "github.com/byuoitav/device-monitoring/actions/then"
)

var uiURL string

func main() {
	go actions.ActionManager().Start(context.TODO())
	messenger.Get().Register(actions.ActionManager().EventStream)

	pflag.StringVar(&uiURL, "ui-url", "", "url to redirect to the ui")
	pflag.Parse()
	// subscribe to something?

	// server
	port := ":10000"
	router := common.NewRouter()

	// remove this eventually
	// redirect from /dash to /dashboard
	router.GET("/dash", func(context echo.Context) error {
		return context.Redirect(http.StatusMovedPermanently, "/dashboard")
	})

	router.GET("/", func(context echo.Context) error {
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
	router.GET("/device/dhcp", handlers.GetDHCPState)
	router.GET("/device/screenshot", handlers.GetScreenshot)
	router.GET("/device/hardwareinfo", handlers.HardwareInfo)
	router.PUT("/device/health", handlers.GetServiceHealth)

	// room info endpoints
	router.GET("/room/ping", handlers.PingRoom)
	router.GET("/room/state", handlers.RoomState)
	router.GET("/room/activesignal", handlers.ActiveSignal)
	router.GET("/room/hardwareinfo", handlers.DeviceHardwareInfo)
	router.GET("/room/viainfo", handlers.ViaInfo)
	router.GET("/room/health", handlers.RoomHealth)

	// action endpoints
	router.PUT("/device/reboot", handlers.RebootPi)
	router.PUT("/device/dhcp/:state", handlers.SetDHCPState)
	router.POST("/event", handlers.SendEvent)

	// divider sensors
	router.GET("/divider/state", handlers.GetDividerState)
	router.GET("/divider/preset/:hostname", handlers.PresetForHostname)

	// flush dns cache
	router.GET("/dns", handlers.FlushDNS)

	router.GET("/ui", redirectHandler)

	/*
		// test mode endpoints
		// router.GET("/maintenance", handlers.IsInMaintMode)
		// router.PUT("/maintenance", handlers.ToggleMaintMode)

		// provisioning endpoints
		router.GET("/provisioning/ws", socket.UpgradeToWebsocket(provisioning.SocketManager()))
		router.GET("/provisioning/id", handlers.GetProvisioningID)
	*/

	router.GET("/actions", actions.ActionManager().Info)
	router.GET("/actions/trigger/:trigger", actions.ActionManager().Config.ActionsByTrigger)

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}
	router.StartServer(&server)
}

func redirectHandler(ctx echo.Context) error {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("error getting hostname: %v", err)
	}

	if uiURL == "" {
		return ctx.Redirect(http.StatusTemporaryRedirect, "http://"+hostname+"/")
	}
	return ctx.Redirect(http.StatusTemporaryRedirect, "http://"+uiURL)
}
