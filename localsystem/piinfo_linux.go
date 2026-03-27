//go:build linux

package localsystem

import (
	"bufio"
	"os"
	"strings"
)

// piModelInfo reads the Raspberry Pi hardware model and OS codename from the
// Linux device tree and /etc/os-release respectively.
func piModelInfo() (model, codename string) {
	// /proc/device-tree/model contains e.g. "Raspberry Pi 4 Model B Rev 1.4\x00"
	if b, err := os.ReadFile("/proc/device-tree/model"); err == nil {
		model = strings.TrimRight(strings.TrimSpace(string(b)), "\x00")
	}

	// /etc/os-release contains VERSION_CODENAME=bookworm (or quoted)
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "VERSION_CODENAME=") {
			codename = strings.Trim(strings.TrimPrefix(line, "VERSION_CODENAME="), "\"")
			break
		}
	}
	return
}
