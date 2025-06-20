package screenshot

import (
	"bytes"
	"context"
	"os"
	"os/exec"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
)

// Take captures a screenshot of the current display and returns it as a byte slice.
func Take(ctx context.Context) ([]byte, *nerr.E) {
	// needed to be fixed to use grim since xwdump is deprecated
	log.L.Infof("Taking screenshot of the Pi using grim")

	xdg := os.Getenv("XDG_RUNTIME_DIR")
	wayland := os.Getenv("WAYLAND_DISPLAY")

	if xdg == "" || wayland == "" {
		log.L.Warnf("Environment not ready: XDG_RUNTIME_DIR=%q, WAYLAND_DISPLAY=%q", xdg, wayland)
		return nil, nerr.Create("Wayland environment variables not set; cannot take screenshot", "wayland-env-missing")
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

	log.L.Debugf("Running grim screenshot command: %v", cmd.Args)

	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return nil, nerr.Translate(err).Addf("unable to take screenshot: %s", stderr.String())
		}
		return nil, nerr.Translate(err).Addf("unable to take screenshot")
	}

	log.L.Debugf("Successfully took screenshot. Size: %d bytes", out.Len())
	return out.Bytes(), nil
}
