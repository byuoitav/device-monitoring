package pi

import (
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
)

// Hostname returns the hostname of the device
func Hostname() (string, *nerr.E) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", nerr.Translate(err).Addf("failed to get hostname.")
	}

	return hostname, nil
}

// IPAddress gets the public ip address of the device
func IPAddress() (net.IP, *nerr.E) {
	var ip net.IP
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, nerr.Translate(err).Addf("failed to get ip address of device")
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && strings.Contains(address.String(), "/24") {
			ip, _, err = net.ParseCIDR(address.String())
			if err != nil {
				return nil, nerr.Translate(err).Addf("failed to get ip address of device")
			}
		}
	}

	if ip == nil {
		return nil, nerr.Create("failed to get ip address of device", "string")
	}

	log.L.Infof("My IP address is %v", ip.String())
	return ip, nil
}

// IsConnectedToInternet returns true if the device can reach google's servers.
func IsConnectedToInternet() bool {
	_, err := net.Dial("tcp", "google.com:80")
	if err != nil {
		return false
	}

	return true
}

// Reboot reboots the device.
func Reboot() {
	log.L.Warnf("*!!* REBOOTING DEVICE NOW *!!*")
	syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}
