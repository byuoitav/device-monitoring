package handlers

import (
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
)

func FlushDNS(c *gin.Context) {
	cmd := exec.Command("sudo", "systemctl", "restart", "dnsmasq")
	err := cmd.Run()
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to flush DNS: %v", err)
		return
	}
	c.String(http.StatusOK, "DNS flushed successfully")
}
