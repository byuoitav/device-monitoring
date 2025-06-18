//go:build linux || darwin
// +build linux darwin

package localsystem

import (
	"fmt"
	"log/slog"

	"golang.org/x/sys/unix"
)

// Reboot reboots the device.
func Reboot() error {
	slog.Info("*!!* REBOOTING DEVICE NOW *!!*")

	err := unix.Reboot(unix.LINUX_REBOOT_CMD_RESTART)
	if err != nil {
		slog.Error("failed to reboot device", slog.Any("error", err))
		return fmt.Errorf("failed to reboot device: %w", err)
	}

	return nil
}
