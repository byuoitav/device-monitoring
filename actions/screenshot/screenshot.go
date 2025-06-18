package screenshot

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
)

// Take -> moving this to wayland since x11 is deprecated
func Take(ctx context.Context) ([]byte, error) {
	slog.Info("Taking screenshot of the pi")

	// get the xwd dump
	xwd := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := exec.Command("/usr/bin/xwd", "-root", "-display", ":0")
	cmd.Stdout = xwd
	cmd.Stderr = stderr
	// cmd.Env = []string{"DISPLAY=:0"}

	slog.Debug("Getting xwd screenshot with command", slog.String("command", cmd.String()))

	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return []byte{}, fmt.Errorf("unable to take a screenshot: %s", stderr)
		}
		return []byte{}, fmt.Errorf("unable to take a screenshot: %w", err)
	}

	// convert the xwd dump to a jpg
	jpg := &bytes.Buffer{}
	cmd = exec.Command("/usr/bin/convert", "xwd:-", "jpg:-")
	cmd.Stdin = xwd
	cmd.Stdout = jpg
	cmd.Stderr = stderr

	slog.Debug("Converting xwd screenshot to jpg with command", slog.String("command", cmd.String()))
	err = cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return []byte{}, fmt.Errorf("unable to take screenshot: %s", stderr)
		}

		return []byte{}, fmt.Errorf("unable to take screenshot: %w", err)
	}
	slog.Debug("Successfully took screenshot", slog.String("size", jpg.String()))
	return jpg.Bytes(), nil
}
