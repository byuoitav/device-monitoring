package handlers

import (
	"net/http"
	"time"

	"log/slog"

	"github.com/byuoitav/device-monitoring/localsystem"
	"github.com/gin-gonic/gin"
)

// RebootPi reboots the pi
func RebootPi(c *gin.Context) {
	go func() {
		for i := 5; i > 0; i-- {
			slog.Info("Reboot countdown", slog.Int("seconds_remaining", i))
			time.Sleep(1 * time.Second)
		}
		if err := localsystem.Reboot(); err != nil {
			slog.Error("Failed to reboot pi", slog.Any("error", err))
		}
	}()
	c.Data(http.StatusOK, "text/plain", []byte("Rebooting in 5 seconds..."))
}

// SetDHCPState toggles dhcp to be on/off
func SetDHCPState(c *gin.Context) {
	state := c.Param("state")
	slog.Info("Received request to set DHCP state", slog.String("state", state))

	// TODO: implement actual DHCP-toggle logic here
	c.String(http.StatusNotImplemented, "SetDHCPState is not yet implemented")
}
