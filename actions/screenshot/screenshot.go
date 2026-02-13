package screenshot

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

// Take -> moving this to wayland since x11 is deprecated
func Take(ctx context.Context) ([]byte, error) {
	slog.Info("Taking screenshot of the Pi using grim")

	xdg := os.Getenv("XDG_RUNTIME_DIR")
	wayland := os.Getenv("WAYLAND_DISPLAY")

	if xdg == "" || wayland == "" {
		slog.Warn("Environment not ready", slog.String("XDG_RUNTIME_DIR", xdg), slog.String("WAYLAND_DISPLAY", wayland))
		return nil, fmt.Errorf("environment not ready: XDG_RUNTIME_DIR=%q, WAYLAND_DISPLAY=%q", xdg, wayland)
	}

	// Detect active output
	output, err := detectActiveOutput(ctx)
	if err != nil {
		slog.Error("Failed to detect active output", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to detect active output: %w", err)
	}

	var cmd *exec.Cmd
	if output != "" {
		slog.Info("Detected active output", slog.String("output", output))
		cmd = exec.CommandContext(ctx, "/usr/bin/grim", "-o", output, "-")
	} else {
		slog.Warn("No output detected, falling back to full screen capture")
		cmd = exec.CommandContext(ctx, "/usr/bin/grim", "-")
	}

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Pass necessary Wayland env variables explicitly
	cmd.Env = append(os.Environ(),
		"XDG_RUNTIME_DIR="+xdg,
		"WAYLAND_DISPLAY="+wayland,
	)

	slog.Debug("Running grim screenshot command", slog.Any("args", cmd.Args))

	err = cmd.Run()
	if err != nil {
		slog.Error("Failed to take screenshot", slog.String("error", err.Error()), slog.String("stderr", stderr.String()))
		return nil, fmt.Errorf("failed to take screenshot: %w, stderr: %s", err, stderr.String())
	}

	slog.Info("Screenshot taken successfully", slog.Int("size_bytes", out.Len()))
	return out.Bytes(), nil
}

// detectActiveOutput detects the active output in a Wayland session.
func detectActiveOutput(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "wlr-randr")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run wlr-randr: %w", err)
	}

	lines := strings.Split(out.String(), "\n")
	for i := range lines {
		line := lines[i]
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, " ") {
			continue
		}

		// Line is the output header: e.g., "DSI-1 \"(null) (null) (DSI-1)\""
		outputName := strings.Fields(line)[0]

		// Look ahead for "Enabled: yes"
		for j := i + 1; j < len(lines); j++ {
			subLine := lines[j]
			if !strings.HasPrefix(subLine, " ") {
				break // next block
			}
			if strings.Contains(subLine, "Enabled: yes") {
				return outputName, nil
			}
		}
	}

	return "", fmt.Errorf("no enabled output found")
}
