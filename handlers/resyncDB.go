package handlers

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os/exec"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
)

// proxy-free, bounded client for internal calls
var internalHTTP = &http.Client{
	Timeout: 15 * time.Second,
	Transport: &http.Transport{
		Proxy: nil, // ignore HTTP(S)_PROXY env vars
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          10,
		IdleConnTimeout:       30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

func ResyncDB(ctx *gin.Context) {
	swabBase := "http://127.0.0.1:7012"
	uiAPIBase := "http://127.0.0.1:8888"
	rplURL := swabBase + "/replication/start"
	uiURL := uiAPIBase + "/refresh"

	// Kick off replication with its own context (not tied to the client)
	internalCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	slog.Info("Starting database replication", slog.String("url", rplURL))
	status, body, err := do(internalCtx, http.MethodGet, rplURL)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			slog.Error("Replication request timed out", slog.String("url", rplURL))
		} else {
			slog.Error("Replication call failed", slog.String("error", err.Error()))
		}
		ctx.String(http.StatusInternalServerError, "replication call failed: %v", err)
		return
	}

	// Accept any 2xx and 409=already running
	if status/100 != 2 && status != http.StatusConflict {
		slog.Error("Replication returned non-success",
			slog.Int("status", status),
			slog.String("body", body),
		)
		ctx.String(http.StatusInternalServerError, "failed to start replication: %d %s", status, body)
		return
	}

	// Optional tiny wait
	select {
	case <-time.After(2 * time.Second):
	case <-ctx.Request.Context().Done():
		// client left; continue best-effort anyway
	}

	// UI refresh
	if uiURL != "" {
		if s2, b2, err2 := do(internalCtx, http.MethodPut, uiURL); err2 != nil {
			slog.Warn("UI refresh call errored", slog.String("error", err2.Error()))
		} else if s2/100 != 2 {
			slog.Warn("UI refresh returned non-2xx", slog.Int("status", s2), slog.String("body", b2))
		} else {
			slog.Info("UI refresh OK", slog.Int("status", s2))
		}
	}

	// Reply to client first
	ctx.String(http.StatusAccepted, "Replication scheduled")

	// Restart asynchronously so the response isn’t killed mid-flight
	go func() {
		time.Sleep(300 * time.Millisecond) // let response flush

		out, err := exec.Command("sudo", "-n", "systemctl", "restart", "device-monitoring.service").CombinedOutput()
		if err != nil {
			slog.Error("Failed to restart device-monitoring.service",
				slog.String("error", err.Error()),
				slog.String("output", string(out)),
			)
			return
		}
		color.Green("Resynced DB / service restarted: %s", string(out))
	}()
}

// do performs an internal HTTP call with proxy-free transport and returns status + up to 4KB body
func do(ctx context.Context, method, url string) (int, string, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return 0, "", err
	}
	resp, err := internalHTTP.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
	return resp.StatusCode, string(b), nil
}
