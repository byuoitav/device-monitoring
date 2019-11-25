// +build linux darwin

package localsystem

import (
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"golang.org/x/sys/unix"
)

// Reboot reboots the device.
func Reboot() *nerr.E {
	log.L.Infof("*!!* REBOOTING DEVICE NOW *!!*")

	err := unix.Reboot(unix.LINUX_REBOOT_CMD_RESTART)
	if err != nil {
		return nerr.Translate(err).Addf("failed to reboot device")
	}

	return nil
}
