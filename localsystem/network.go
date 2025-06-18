package localsystem

import (
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	dhcpFile = "/etc/dhcpcd.conf"
)

// Hostname returns the hostname of the device
func Hostname() (string, error) {
	h, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}
	return h, nil
}

// MustHostname returns the hostname of the device, and panics if it fails
func MustHostname() string {
	h, err := Hostname()
	if err != nil {
		log.Fatalf("failed to get hostname: %v", err)
	}
	return h
}

// IPAddress gets the public ip address of the device
func IPAddress() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("failed to list network interfaces: %w", err)
	}

	var ip net.IP
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok &&
			!ipnet.IP.IsLoopback() &&
			strings.Contains(addr.String(), "/24") {

			parsed, _, perr := net.ParseCIDR(addr.String())
			if perr != nil {
				return nil, fmt.Errorf("failed to parse IP %q: %w", addr.String(), perr)
			}
			ip = parsed
			break
		}
	}

	if ip == nil {
		return nil, fmt.Errorf("no non‑loopback /24 IP address found")
	}

	slog.Info("My IP address", slog.String("ip", ip.String()))
	return ip, nil
}

// IsConnectedToInternet returns true if the device can reach google's servers.
func IsConnectedToInternet() bool {
	conn, err := net.Dial("tcp", "google.com:80")
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// UsingDHCP returns true if the device is using DHCP, and false if it has a static ip set.tenidno
func UsingDHCP() (bool, error) {
	// read dhcpcd.conf file
	contents, err := os.ReadFile(dhcpFile)
	if err != nil {
		return false, fmt.Errorf("unable to read %s: %w", dhcpFile, err)
	}

	staticRE := regexp.MustCompile(`(?m)^static ip_address`)
	if staticRE.Match(contents) {
		return false, nil
	}
	return true, nil
}

// ToggleDHCP turns dhcp on/off by swapping dhcpcd.conf with dhcpcd.conf.other, a file we created when the pi was setup.
func ToggleDHCP() error {
	// validate the necessary files exist
	if err := CanToggleDHCP(); err != nil {
		return err
	}

	tmp := dhcpFile + ".tmp"
	other := dhcpFile + ".other"

	if err := os.Rename(dhcpFile, tmp); err != nil {
		return fmt.Errorf("rename %s→%s: %w", dhcpFile, tmp, err)
	}
	if err := os.Rename(other, dhcpFile); err != nil {
		return fmt.Errorf("rename %s→%s: %w", other, dhcpFile, err)
	}
	if err := os.Rename(tmp, other); err != nil {
		return fmt.Errorf("rename %s→%s: %w", tmp, other, err)
	}

	// restart dhcpcd
	out, err := exec.Command("sh", "-c", "sudo systemctl restart dhcpcd").CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart dhcpcd: %w — %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// CanToggleDHCP returns nil if you can toggle DHCP, or an error if you can't
func CanToggleDHCP() error {
	if _, err := os.Stat(dhcpFile); os.IsNotExist(err) {
		return fmt.Errorf("%s does not exist; cannot toggle DHCP", dhcpFile)
	}
	other := dhcpFile + ".other"
	if _, err := os.Stat(other); os.IsNotExist(err) {
		return fmt.Errorf("%s does not exist; cannot toggle DHCP", other)
	}
	return nil
}
