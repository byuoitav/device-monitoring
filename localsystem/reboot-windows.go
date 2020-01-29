// +build windows

package localsystem

import (
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
)

// Reboot reboots the device.
func Reboot() *nerr.E {
	log.L.Infof("*!!* REBOOTING DEVICE NOW *!!*")

	return nerr.Createf("Error", "cannot reboot windows")

}
