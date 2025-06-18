//go:build windows
// +build windows

package localsystem

import (
	"fmt"
	"log/slog"
	"os/exec"
)

// Reboot reboots the device.
func Reboot() error {
	slog.Info("*!!* REBOOTING DEVICE NOW *!!*")
	// Windows does not support rebooting through the golang standard library.
	// Instead, we can use the "shutdown" command with the "/r" flag to reboot.
	cmd := "shutdown /r /t 0"
	_, err := exec.Command("cmd", "/C", cmd).Output()
	if err != nil {
		slog.Error("failed to reboot device", slog.Any("error", err))
		return fmt.Errorf("failed to reboot device: %w", err)
	}
	// If we reach here, the command was successful.
	slog.Info("reboot command executed successfully")
	return nil
}
