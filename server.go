package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/byuoitav/auth/wso2"
	"github.com/byuoitav/device-monitoring/actions"
	"github.com/byuoitav/device-monitoring/couchdb"
	"github.com/byuoitav/device-monitoring/handlers"
	"github.com/byuoitav/device-monitoring/messenger"
	"github.com/byuoitav/device-monitoring/model"
	"github.com/gin-gonic/gin"
	"github.com/labstack/echo"
	"github.com/spf13/pflag"

	"github.com/byuoitav/device-monitoring/actions/gpio"
	"github.com/lmittmann/tint"

	_ "github.com/byuoitav/device-monitoring/actions/then"
)

func main() {
	// ===========================
	// Flags (PRD required; STG optional)
	// ===========================
	// Flight-Deck REST API bases
	prdAPIBase := pflag.String("fd-prd-api-base", "https://api.byu.edu/domains/av/flight-deck/v2", "Flight-Deck PRD API base")
	stgAPIBase := pflag.String("fd-stg-api-base", "", "Flight-Deck STG API base (optional fallback)")

	// WSO2 issuer/gateway + OAuth client per env
	prdGatewayURL := pflag.String("prd-gateway-url", "", "WSO2 PRD gateway URL (issuer/gateway base, e.g. https://api.byu.edu)")
	prdClientID := pflag.String("prd-client-id", "", "WSO2 PRD client id")
	prdClientSecret := pflag.String("prd-client-secret", "", "WSO2 PRD client secret")

	stgGatewayURL := pflag.String("stg-gateway-url", "", "WSO2 STG gateway URL")
	stgClientID := pflag.String("stg-client-id", "", "WSO2 STG client id")
	stgClientSecret := pflag.String("stg-client-secret", "", "WSO2 STG client secret")

	pflag.Parse()

	// ---- Validate PRD flags (required) ----
	if *prdGatewayURL == "" || *prdClientID == "" || *prdClientSecret == "" || *prdAPIBase == "" {
		slog.Error("Missing required PRD flags. Provide: --prd-gateway-url --prd-client-id --prd-client-secret --fd-prd-api-base")
		os.Exit(1)
	}

	// ===========================
	// Logging
	// ===========================
	w := os.Stderr
	handler := tint.NewHandler(w, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.Kitchen,
		NoColor:    false,
	})
	slog.SetDefault(slog.New(handler))
	slog.Info("Starting device-monitoring server")

	// ===========================
	// External deps checks
	// ===========================
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := couchdb.ValidateConnection(ctx); err != nil {
		slog.Error("Failed to connect to CouchDB", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("Successfully connected to CouchDB")

	// ===========================
	// Initialize Flight-Deck clients for RefreshContainers
	// ===========================
	fdCfg := handlers.FDConfig{
		PRD: handlers.FDEnvConfig{
			APIBase:      *prdAPIBase,    // e.g. https://api.byu.edu/domains/av/flight-deck/v2
			GatewayURL:   *prdGatewayURL, // e.g. https://api.byu.edu
			ClientID:     *prdClientID,
			ClientSecret: *prdClientSecret,
			// Scopes: []string{"flight-deck.refloat"}, // uncomment if your wso2.Client needs explicit scopes
		},
	}
	// Wire STG only if all four provided
	if *stgAPIBase != "" && *stgGatewayURL != "" && *stgClientID != "" && *stgClientSecret != "" {
		fdCfg.STG = handlers.FDEnvConfig{
			APIBase:      *stgAPIBase,
			GatewayURL:   *stgGatewayURL,
			ClientID:     *stgClientID,
			ClientSecret: *stgClientSecret,
		}
	}

	if err := handlers.InitFlightDeck(fdCfg); err != nil {
		slog.Error("InitFlightDeck failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Keep legacy global WSO2 client (PRD) for other handlers that still reference handlers.WSO2Client
	handlers.WSO2Client = *wso2.New(*prdClientID, *prdClientSecret, *prdGatewayURL, "device-monitoring")

	pins, err := loadPinsFromJSON(couchdb.GetMonitoringConfig(ctx))
	if err != nil {
		slog.Error("Failed to load GPIO pins from JSON", slog.Any("error", err))
	} else {
		gpio.SetPins(pins)
		gpio.StartAllMonitors()
		time.Sleep(250 * time.Millisecond) // give monitors a moment to start and read initial states
		slog.Info("GPIO monitors started", slog.Int("pinCount", len(pins)))
	}

	fmt.Println(pins)

	// ===========================
	// Start action manager
	// ===========================
	go actions.ActionManager().Start(context.TODO())
	messenger.Get().Register(model.ChanEventConverter(actions.ActionManager().EventStream))

	// ===========================
	// Gin router
	// ===========================
	port := ":10000"
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// redirects
	router.GET("/dash", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/dashboard") })
	router.GET("/", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/dashboard") })

	// static dashboard
	router.Static("/dashboard", "./dashboard")

	// SPA fallback under /dashboard
	router.NoRoute(func(c *gin.Context) {
		if len(c.Request.URL.Path) >= 10 && c.Request.URL.Path[:10] == "/dashboard" {
			c.File(filepath.Join("dashboard", "index.html"))
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
	})

	// health
	router.GET("/ping", func(c *gin.Context) {
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

	// room info endpoints
	router.GET("/room/ping", handlers.PingRoom)
	router.GET("/room/state", handlers.RoomState)
	router.GET("/room/activesignal", handlers.ActiveSignal)
	router.GET("/room/hardwareinfo", handlers.DeviceHardwareInfo)
	router.GET("/room/health", handlers.RoomHealth)

	// actions
	router.PUT("/device/reboot", handlers.RebootPi)
	router.PUT("/device/dhcp/:state", handlers.SetDHCPState)
	router.POST("/event", handlers.SendEvent)

	// divider sensors
	router.GET("/divider/state", handlers.GetDividerState)
	router.GET("/divider/preset/:hostname", handlers.PresetForHostname)

	// action manager
	router.GET("/actions", func(c *gin.Context) { c.JSON(http.StatusOK, echoToGin(actions.ActionManager().Info)) })
	router.GET("/actions/trigger/:trigger", func(c *gin.Context) {
		c.JSON(http.StatusOK, echoToGin(actions.ActionManager().Config.ActionsByTrigger))
	})

	// utilities
	router.GET("/dns", handlers.FlushDNS)
	router.GET("/resyncDB", handlers.ResyncDB)
	router.GET("/refreshContainers", handlers.RefreshContainers) // hardened handler

	// API group
	api := router.Group("/api")
	api.GET("/v1/monitoring", handlers.GetDeviceHealth)

	// run!
	router.Run(port)
}

// echoToGin adapts an Echo handler to a Gin handler (this is hacky and needs to be fixed)
func echoToGin(eh func(echo.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		eCtx := echo.New().NewContext(c.Request, c.Writer)
		for _, param := range c.Params {
			eCtx.SetParamNames(param.Key)
			eCtx.SetParamValues(param.Value)
		}
		if err := eh(eCtx); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		}
	}
}

// loadPinsFromJSON loads GPIO pin configurations from a couchDoc
func loadPinsFromJSON(cfg map[string]any, err error) ([]gpio.Pin, error) {
	if err != nil {
		return nil, fmt.Errorf("failed to get couch config: %w", err)
	}

	// Marshal the generic map back to JSON so we can unmarshal into
	// strongly-typed structs that include []Pin.
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal couch config: %w", err)
	}

	// Shape of the CouchDB document, but we only care about the parts
	// that lead to the pin configuration.
	type couchDoc struct {
		Actions []struct {
			Name string `json:"name"`
			Then []struct {
				Do   string     `json:"do"`
				With []gpio.Pin `json:"with"`
			} `json:"then"`
		} `json:"actions"`
	}

	var doc couchDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal couch config: %w", err)
	}

	var pins []gpio.Pin

	// Find the "monitor-dividers" action (the one that contains the pin config)
	for _, action := range doc.Actions {
		if action.Name != "monitor-dividers" {
			continue
		}

		for _, step := range action.Then {
			pins = append(pins, step.With...)
		}
	}

	if len(pins) == 0 {
		return nil, fmt.Errorf("no pin configuration found in couch config")
	}

	return pins, nil
}
