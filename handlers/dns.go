package handlers

import (
	"net/http"
	"os/exec"

	"github.com/labstack/echo"
)

func FlushDNS(c echo.Context) error {
	cmd := exec.Command("sudo", "systemctl", "restart", "dnsmasq")
	err := cmd.Run()
	if err != nil {
		return c.String(http.StatusInternalServerError, "failure")
	}
	return c.String(http.StatusOK, "success")
}
