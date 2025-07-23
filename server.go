package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/byuoitav/device-monitoring/actions"
	"github.com/byuoitav/device-monitoring/couchdb"
	"github.com/byuoitav/device-monitoring/handlers"
	"github.com/byuoitav/device-monitoring/messenger"
	"github.com/byuoitav/device-monitoring/model"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"

	_ "github.com/byuoitav/device-monitoring/actions/then"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := couchdb.ValidateConnection(ctx); err != nil {
		slog.Error("Failed to connect to CouchDB", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("Successfully connected to CouchDB")

	// start the action manager
	go actions.ActionManager().Start(context.TODO())
	messenger.Get().Register(model.ChanEventConverter(actions.ActionManager().EventStream))

	// create Gin router
	port := ":10000"
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// redirect from /dash to /dashboard
	router.GET("/dash", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/dashboard")
	})

	// static webpages
	router.Use(static.Serve("/dashboard", static.LocalFile("dashboard", true)))

	// health endpoint
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "Don't meddle in the affairs of Wizards, for they are subtle and quick to anger.")
	})

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

	// room info endpoints]
	// TODO: fix nothing shows on hardwareinfo and health
	router.GET("/room/ping", handlers.PingRoom)
	router.GET("/room/state", handlers.RoomState)
	router.GET("/room/activesignal", handlers.ActiveSignal)
	router.GET("/room/hardwareinfo", handlers.DeviceHardwareInfo) // nothing shows
	router.GET("/room/health", handlers.RoomHealth)               // nothing shows

	// action endpoints
	// TODO: check is this work with shipwright
	router.PUT("/device/reboot", handlers.RebootPi)
	router.PUT("/device/dhcp/:state", handlers.SetDHCPState)
	router.POST("/event", handlers.SendEvent)

	// divider sensors
	// TODO: send actual feedback message
	router.GET("/divider/state", handlers.GetDividerState)
	router.GET("/divider/preset/:hostname", handlers.PresetForHostname)

	// action manager
	router.GET("/actions", func(c *gin.Context) {
		c.JSON(http.StatusOK, actions.ActionManager().Info)
	})
	router.GET("/actions/trigger/:trigger", func(c *gin.Context) {
		c.JSON(http.StatusOK, actions.ActionManager().Config.ActionsByTrigger)
	})

	// flush dns cache
	router.GET("/dns", handlers.FlushDNS)

	// reSyncDB (old SWAB)
	// TODO: fix this, it was failing with 405 Method Not Allowed
	router.GET("/resyncDB", handlers.ResyncDB) // failed to refresh Ui: 405 Method Not Allowed

	// refreshContainers (old refloat)
	// TODO: fix this, it was response from flight-deck {"error":"Not Authorized"}
	router.GET("/refreshContainers", handlers.RefreshContainers) //{"error":"Not Authorized"} response from flight-deck

	// New Router Group for the API with versioning /api/v1 or /api/v2 etc.
	// This is where you would add your API endpoints
	api := router.Group("/api")
	// returns JSON of all the devices and their health
	// TODO: fix missing room_id
	api.GET("/v1/monitoring", handlers.GetDeviceHealth) // missing room_id

	// run!
	router.Run(port)
}

// echoToGinHandler adapts an Echo handler to a Gin handler
/*func echoToGinHandler(echoHandler func(echo.Context) error) gin.HandlerFunc {
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
}*/
