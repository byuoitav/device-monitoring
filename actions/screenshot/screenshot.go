package screenshot

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

// Take -> moving this to wayland since x11 is deprecated
func Take(ctx context.Context) ([]byte, error) {
	slog.Info("Taking screenshot of the Pi using grim")

	xdg := os.Getenv("XDG_RUNTIME_DIR")
	wayland := os.Getenv("WAYLAND_DISPLAY")

	if xdg == "" || wayland == "" {
		slog.Warn("Environment not ready: XDG_RUNTIME_DIR=%q, WAYLAND_DISPLAY=%q", xdg, wayland)
		return nil, fmt.Errorf("environment not ready: XDG_RUNTIME_DIR=%q, WAYLAND_DISPLAY=%q", xdg, wayland)
	}

	var out bytes.Buffer
	var stderr bytes.Buffer

	// Use grim with stdout to capture the entire screen
	cmd := exec.CommandContext(ctx, "/usr/bin/grim", "-o", "DSI-1", "-") // "-" = write to stdout
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// explicitly pass the env
	cmd.Env = append(os.Environ(),
		"XDG_RUNTIME_DIR="+xdg,
		"WAYLAND_DISPLAY="+wayland,
	)

	slog.Debug("Running grim screenshot command", slog.Any("command", cmd.String()))

	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			slog.Error("Failed to take screenshot", slog.String("error", err.Error()), slog.String("stderr", stderr.String()))
			return nil, fmt.Errorf("failed to take screenshot: %w, stderr: %s", err, stderr.String())
		}
		slog.Error("Failed to take screenshot", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to take screenshot: %w", err)
	}

	slog.Info("Screenshot taken successfully", slog.Int("size_bytes", out.Len()))
	return out.Bytes(), nil
}
