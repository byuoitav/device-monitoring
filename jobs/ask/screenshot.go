package ask

import (
	"bytes"
	"os/exec"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/v2/events"
)

// ScreenshotJob takes a screenshot of what is currently on the pi
type ScreenshotJob struct{}

// Run runs the job
func (j *ScreenshotJob) Run(ctx interface{}, eventWrite chan events.Event) interface{} {
	xwd := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	log.L.Infof("Taking screenshot of the pi")

	cmd := exec.Command("XAUTHORITY=/home/pi/.Xauthority /usr/bin/xwd", "-root")
	cmd.Stdout = xwd
	cmd.Stderr = stderr
	cmd.Env = []string{"DISPLAY=:0.0"}

	log.L.Debugf("Getting xwd screenshot with command: %s", cmd.Args)
	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return nerr.Translate(err).Addf("failed to get a screenshot: %s", stderr)
		}

		return nerr.Translate(err).Addf("failed to get a screenshot")
	}

	jpg := &bytes.Buffer{}
	cmd = exec.Command("/usr/bin/convert", "xwd:-", "jpg:-")
	cmd.Stdin = xwd
	cmd.Stdout = jpg
	cmd.Stderr = stderr

	log.L.Debugf("Converting xwd screenshot to jpeg with command: %s", cmd.Args)
	err = cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return nerr.Translate(err).Addf("failed to get a screenshot: %s", stderr)
		}

		return nerr.Translate(err).Addf("failed to get a screenshot")
	}

	log.L.Debugf("Successfully took screenshot.")
	return jpg
}
