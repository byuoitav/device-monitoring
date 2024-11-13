package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo"
	"golang.org/x/crypto/ssh"
)

// ResyncDB (Swab)
func ResyncDB(ctx echo.Context) error {
	// get the pi's address
	piHostname, err := os.Hostname()
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Errorf("failed to get hostname: %w", err).Error())
	}

	// start the db replication
	req, err := http.NewRequestWithContext(ctx.Request().Context(), http.MethodGet, fmt.Sprintf("http://%s:7012/replication/start", piHostname), nil)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Errorf("failed to create replication request: %w", err).Error())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Errorf("failed to start replication: %w", err).Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ctx.String(http.StatusInternalServerError, fmt.Errorf("failed to start replication: %s", resp.Status).Error())
	}

	// replication waiting
	select {
	case <-time.After(10 * time.Second):
	case <-ctx.Request().Context().Done():
		return ctx.String(http.StatusInternalServerError, "context cancelled")
	}

	// refresh the ui
	req, err = http.NewRequestWithContext(ctx.Request().Context(), http.MethodPut, fmt.Sprintf("http://%s:8888/refresh", piHostname), nil)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Errorf("failed to create refresh ui request: %w", err).Error())
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Errorf("failed to refresh ui: %w", err).Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ctx.String(http.StatusInternalServerError, fmt.Errorf("failed to refresh ui: %s", resp.Status).Error())
	}

	// restart the device monitoring service (dmm)

	return ctx.String(http.StatusOK, "Resyncing DB")
}

// RefreshContainers (Float)
func RefreshContainers(ctx echo.Context) error {
	var req *http.Request
	var err error

	// get the pi's address
	piHostname, err := os.Hostname()
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Errorf("failed to get hostname: %w", err).Error())
	}

	newCtx, cancel := context.WithTimeout(ctx.Request().Context(), 30*time.Second)
	defer cancel()

	req, err = http.NewRequestWithContext(newCtx, http.MethodPost, fmt.Sprintf("https://api.byu.edu/domains/av/flight-deck/v2/refloat/%v", piHostname), nil)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Errorf("failed to create refloat request: %w", err).Error())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Errorf("failed to refloat: %w", err).Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { // 200
		body, err := io.ReadAll(resp.Body)
		if err != nil { // 500
			return ctx.String(http.StatusInternalServerError, fmt.Errorf("%v response from flight-deck", err).Error())
		}

		return ctx.String(http.StatusInternalServerError, fmt.Errorf("%v response from flight-deck", string(body)).Error())
	}

	return ctx.String(http.StatusOK, "Refreshed Containers")
}

// helper ssh into the pi and execute "sudo systemctl restart device-monitoring.service"
func piSSH(c echo.Context) (*ssh.Client, error) {
	// get the pi's address
	piHostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}
	// ssh into the pi
	// execute "sudo systemctl restart device-monitoring.service"
	deadline, ok := c.Request().Context().Deadline()
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
