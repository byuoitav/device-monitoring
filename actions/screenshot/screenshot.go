package screenshot

import (
	"bytes"
	"context"
	"os/exec"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
)

// Take .
func Take(ctx context.Context) ([]byte, *nerr.E) {
	log.L.Infof("Taking screenshot of the pi")

	// get the xwd dump
	xwd := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := exec.Command("/usr/bin/xwd", "-root", "-display", ":0")
	cmd.Stdout = xwd
	cmd.Stderr = stderr
	// cmd.Env = []string{"DISPLAY=:0"}

	log.L.Debugf("Getting xwd screenshot with command: %s", cmd.Args)

	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return []byte{}, nerr.Translate(err).Addf("unable to take a screenshot: %s", stderr)
		}

		return []byte{}, nerr.Translate(err).Addf("unable to take a screenshot")
	}

	// convert the xwd dump to a jpg
	jpg := &bytes.Buffer{}
	cmd = exec.Command("/usr/bin/convert", "xwd:-", "jpg:-")
	cmd.Stdin = xwd
	cmd.Stdout = jpg
	cmd.Stderr = stderr

	log.L.Debugf("Converting xwd screenshot to jpg with command: %s", cmd.Args)
	err = cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return []byte{}, nerr.Translate(err).Addf("unable to take screenshot: %s", stderr)
		}

		return []byte{}, nerr.Translate(err).Addf("unable to take screenshot")
	}

	log.L.Debugf("Successfully took screenshot.")
	return jpg.Bytes(), nil
}
