package handlers

import (
	"net/http"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/device-monitoring/pi"
	"github.com/labstack/echo"
)

// RebootPi reboots the pi
func RebootPi(context echo.Context) error {
	go func() {
		for i := 5; i > 0; i-- {
			log.L.Infof("REBOOTING PI IN %v SECONDS", i)
			time.Sleep(1 * time.Second)
		}

		err := pi.Reboot()
		if err != nil {
			log.L.Errorf("failed to reboot pi: %v", err.Error())
		}
	}()

	return context.Blob(http.StatusOK, "text/plain", []byte("Rebooting in 5 seconds..."))
}

// SetDHCPState toggles dhcp to be on/off
func SetDHCPState(ctx echo.Context) error {
	return nil
}
