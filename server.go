package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/byuoitav/device-monitoring/actions"
	"github.com/byuoitav/device-monitoring/couchdb"
	"github.com/byuoitav/device-monitoring/handlers"
	"github.com/byuoitav/device-monitoring/messenger"
	"github.com/byuoitav/device-monitoring/model"
	"github.com/gin-gonic/gin"
	"github.com/labstack/echo"

	"github.com/lmittmann/tint"

	_ "github.com/byuoitav/device-monitoring/actions/then"
)

func main() {

	// set up logging
	w := os.Stderr
	handler := tint.NewHandler(w, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.Kitchen,
	})
	slog.SetDefault(slog.New(handler))

	slog.Info("Starting device-monitoring server")

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

	router.Static("/dashboard", "./dashboard")

	// HTML5 fallback for SPA (if requested file doesn't exist)
	router.NoRoute(func(c *gin.Context) {
		// if request starts with /dashboard, serve index.html
		if len(c.Request.URL.Path) >= 10 && c.Request.URL.Path[:10] == "/dashboard" {
			c.File(filepath.Join("dashboard", "index.html"))
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
	})

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
	router.GET("/room/health", handlers.RoomHealth)

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
		c.JSON(http.StatusOK, echoToGin(actions.ActionManager().Info))
	})

	router.GET("/actions/trigger/:trigger", func(c *gin.Context) {
		c.JSON(http.StatusOK, echoToGin(actions.ActionManager().Config.ActionsByTrigger))
	})

	// flush dns cache
	router.GET("/dns", handlers.FlushDNS)

	// reSyncDB (old SWAB)
	router.GET("/resyncDB", handlers.ResyncDB)

	// refreshContainers (old refloat)
	// TODO: fix this, it was response from flight-deck {"error":"Not Authorized"}
	router.GET("/refreshContainers", handlers.RefreshContainers) //{"error":"Not Authorized"} response from flight-deck

	// New Router Group for the API with versioning /api/v1 or /api/v2 etc.
	// This is where you would add your API endpoints
	api := router.Group("/api")
	// returns JSON of all the devices and their health
	// TODO: check if we want this to use a ?room_id query parameter or get all devices in the room
	api.GET("/v1/monitoring", handlers.GetDeviceHealth)

	// run!
	router.Run(port)
}

// echoToGin adapts an Echo handler to a Gin handler (this is hacky and needs to be fixed)
// This is a workaround to use Echo handlers in Gin, which is not ideal but works for now.
// Needs to be fixed in shipwright or replaced with Gin handlers.
func echoToGin(eh func(echo.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		// create a minimal Echo context
		eCtx := echo.New().NewContext(c.Request, c.Writer)

		// copy params (optional, if Info() uses them)
		for _, param := range c.Params {
			eCtx.SetParamNames(param.Key)
			eCtx.SetParamValues(param.Value)
		}

		// execute the Echo handler
		if err := eh(eCtx); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		}
	}
}
