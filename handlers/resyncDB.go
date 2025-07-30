package handlers

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"
)

// ResyncDB (Swab)
func ResyncDB(ctx *gin.Context) {
	// get the pi's address
	piHostname, err := os.Hostname()
	if err != nil {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to get hostname: %v", err))
		return
	}

	// build localhost
	localhost := fmt.Sprintf("http://%s.byu.edu", piHostname)

	// build the replication request URL
	rplUrl := fmt.Sprintf("%s:7012/replication/start", localhost)

	// start the db replication
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

// helper ssh into the pi and execute "sudo systemctl restart device-monitoring.service"
func piSSH(ctx *gin.Context) (*ssh.Client, error) {
	// get the pi's address
	piHostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	// ssh into the pi
	// execute "sudo systemctl restart device-monitoring.service"
	deadline, ok := ctx.Request.Context().Deadline()
	if !ok {
		deadline = time.Now().Add(3 * time.Second)
	}

	// create the ssh client
	sshClient := &ssh.ClientConfig{
		User:            "pi",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password("raspberry"),
		},
		Timeout: time.Until(deadline),
	}

	// dial the pi
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", piHostname), sshClient)
	if err != nil {
		return nil, fmt.Errorf("failed to dial pi: %w", err)
	}

	return client, nil
}
