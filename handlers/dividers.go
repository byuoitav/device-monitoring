package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/byuoitav/device-monitoring/actions/gpio"
	"github.com/byuoitav/device-monitoring/couchdb"
	"github.com/gin-gonic/gin"
)

// GetDividerState returns the state of the dividers
func GetDividerState(c *gin.Context) {
	pins := gpio.GetPins()
	resp := map[string][]string{"connected": {}, "disconnected": {}}
	for _, p := range pins {
		if p.Connected {
			resp["connected"] = append(resp["connected"], p.BlueberryPresets)
		} else {
			resp["disconnected"] = append(resp["disconnected"], p.BlueberryPresets)
		}
	}
	c.JSON(http.StatusOK, resp)
}

// PresetForHostname returns the preset that a specific hostname should be on
func PresetForHostname(c *gin.Context) {
	hostname := c.Param("hostname")

	pins := gpio.GetPins()

	if len(pins) != 1 {
		c.String(http.StatusBadRequest, "not supported in this room")
		return
	}
	preset := pins[0].CurrentPreset(hostname)
	c.String(http.StatusOK, preset)
}

// GetDividerPins returns the configured GPIO pin definitions for divider sensors.
func GetDividerPins(c *gin.Context) {
	customSystemID := c.Params.ByName("systemID")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pins, err := LoadPinsFromJSON(couchdb.GetMonitoringConfig(ctx, customSystemID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pins)
}

// LoadPinsFromJSON loads GPIO pin configurations from a couchDoc.
func LoadPinsFromJSON(cfg map[string]any, err error) ([]gpio.Pin, error) {
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

	// Find the "monitor-dividers" action (the one that contains the pin config).
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
