package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/byuoitav/device-monitoring/actions"
	"github.com/byuoitav/device-monitoring/handlers"
	"github.com/byuoitav/device-monitoring/messenger"
	"github.com/byuoitav/device-monitoring/model"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/labstack/echo"
	"github.com/spf13/pflag"

	_ "github.com/byuoitav/device-monitoring/actions/then"
)

var uiURL string

func main() {
	// start the action manager
	go actions.ActionManager().Start(context.TODO())
	messenger.Get().Register(model.ChanEventConverter(actions.ActionManager().EventStream))

	// parse --ui-url
	pflag.StringVar(&uiURL, "ui-url", "", "url to redirect to the ui")
	pflag.Parse()

	// create Gin router
	port := ":10000"
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(cors.Default())

	// health endpoint
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "Don't meddle in the affairs of Wizards, for they are subtle and quick to anger.")
	})

	// dash & root redirects
	router.GET("/dash", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/dashboard")
	})
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/dashboard")
	})

	// serve SPA (static files)
	router.StaticFS("/dashboard", http.Dir("dashboard"))

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

	// dynamic UI redirect
	router.GET("/ui", redirectHandler)

	/*
		// test mode endpoints
		// router.GET("/maintenance", handlers.IsInMaintMode)
		// router.PUT("/maintenance", handlers.ToggleMaintMode)

		// provisioning endpoints
		router.GET("/provisioning/ws", socket.UpgradeToWebsocket(provisioning.SocketManager()))
		router.GET("/provisioning/id", handlers.GetProvisioningID)
	*/

	router.GET("/actions", echoToGinHandler(actions.ActionManager().Info))
	router.GET("/actions/trigger/:trigger", echoToGinHandler(actions.ActionManager().Config.ActionsByTrigger))

	// reSyncDB (old SWAB)
	router.GET("/resyncDB", handlers.ResyncDB)

	// refreshContainers (old refloat)
	router.GET("/refreshContainers", handlers.RefreshContainers)

	// New Router Group for the API with versioning /api/v1 or /api/v2 etc.
	// This is where you would add your API endpoints
	api := router.Group("/api")
	// returns JSON of all the devices and their health
	api.GET("/v1/monitoring", handlers.GetDeviceHealth)

	// run!
	router.Run(port)
}

// redirectHandler handles the redirect to the UI
func redirectHandler(c *gin.Context) {
	if uiURL != "" {
		c.Redirect(http.StatusTemporaryRedirect, "http://"+uiURL)
		return
	}
	host := strings.Split(c.Request.Host, ":")[0]
	c.Redirect(http.StatusTemporaryRedirect, "http://"+host+"/")
}

// echoToGinHandler adapts an Echo handler to a Gin handler
func echoToGinHandler(echoHandler func(echo.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a fake Echo context to satisfy the handler
		e := echo.New()
		req := c.Request
		res := c.Writer

		// Bind the Gin context to an Echo context
		echoCtx := e.NewContext(req, res)

		// Run the Echo handler
		if err := echoHandler(echoCtx); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		}
	}
}
