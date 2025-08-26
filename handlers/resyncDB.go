package handlers

import (
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
)

// ResyncDB (Swab)
func ResyncDB(ctx *gin.Context) {
	// to resync the database, we need to perform the replication
	// curl -X GET http://localhost:7012/replication/start
	localhost := "http://localhost:7012"
	rplUrl := fmt.Sprintf("%s/replication/start", localhost)

	// create the request
	req, err := http.NewRequestWithContext(ctx.Request.Context(), http.MethodGet, rplUrl, nil)
	if err != nil {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to create replication request: %v", err))
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to start replication: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to start replication: %s", resp.Status))
		return
	}

	// replication waiting
	select {
	case <-time.After(10 * time.Second):
	case <-ctx.Request.Context().Done():
		ctx.String(http.StatusInternalServerError, "context cancelled")
		return
	}

	// refresh UI URL
	uiUrl := fmt.Sprintf("%s:8888/refresh", localhost)
	// refresh the UI
	req, err = http.NewRequestWithContext(ctx.Request.Context(), http.MethodPut, uiUrl, nil)
	if err != nil {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to create refresh UI request: %v", err))
		return
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to refresh UI: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to refresh UI: %s", resp.Status))
		return
	}

	// restart the device monitoring service (dmm) with systemctl
	// after this the dmm goes offliine for a few seconds so we need to wait for it to come back online
	cmd := exec.Command("sudo", "systemctl", "restart", "device-monitoring.service")
	output, err := cmd.Output()
	if err != nil {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to restart device-monitoring.service: %v", err))
		return
	}

	color.Green("Resynced DB: %s", string(output))

	ctx.String(http.StatusOK, "Resyncing DB")
}
