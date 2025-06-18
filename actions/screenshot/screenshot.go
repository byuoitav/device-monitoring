package screenshot

import (
	"bytes"
	"context"
	"os/exec"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
)

// Take captures a screenshot of the current display and returns it as a byte slice.
func Take(ctx context.Context) ([]byte, *nerr.E) {
	// needed to be fixed to use grim since xwdump is deprecated
	log.L.Infof("Taking screenshot of the Pi using grim")

	var out bytes.Buffer
	var stderr bytes.Buffer

	// Use grim with stdout to capture the entire screen
	cmd := exec.CommandContext(ctx, "/usr/bin/grim", "-") // "-" outputs to stdout
	cmd.Stdout = &out
	cmd.Stderr = &stderr

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
