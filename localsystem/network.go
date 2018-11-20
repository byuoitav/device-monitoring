package localsystem

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
)

const (
	dhcpFile = "/etc/dhcpcd.conf"
)

// Hostname returns the hostname of the device
func Hostname() (string, *nerr.E) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", nerr.Translate(err).Addf("failed to get hostname.")
	}

	return hostname, nil
}

// MustHostname returns the hostname of the device, and panics if it fails
func MustHostname() string {
	hostname, err := Hostname()
	if err != nil {
		log.L.Fatalf("failed to get hostname: %s", err.Error())
	}

	return hostname
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

// UsingDHCP returns true if the device is using DHCP, and false if it has a static ip set.
func UsingDHCP() (bool, *nerr.E) {
	// read dhcpcd.conf file
	contents, err := ioutil.ReadFile(dhcpFile)
	if err != nil {
		return false, nerr.Translate(err).Addf("unable to read %s", dhcpFile)
	}

	reg := regexp.MustCompile(`(?m)^static ip_address`)
	matches := reg.Match(contents)

	return !matches, nil
}

// ToggleDHCP turns dhcp on/off by swapping dhcpcd.conf with dhcpcd.conf.other, a file we created when the pi was setup.
func ToggleDHCP() *nerr.E {
	// validate the necessary files exist
	if err := CanToggleDHCP(); err != nil {
		return err
	}

	tmpFile := fmt.Sprintf("%s.tmp", dhcpFile)
	otherFile := fmt.Sprintf("%s.other", dhcpFile)

	// swap the files
	err := os.Rename(dhcpFile, tmpFile)
	if err != nil {
		return nerr.Translate(err)
	}

	err = os.Rename(otherFile, dhcpFile)
	if err != nil {
		return nerr.Translate(err)
	}

	err = os.Rename(tmpFile, otherFile)
	if err != nil {
		return nerr.Translate(err)
	}

	// restart dhcp service
	_, err = exec.Command("sh", "-c", "sudo systemctl restart dhcpcd").Output()
	if err != nil {
		return nerr.Translate(err).Addf("unable to restart dhcpcd service")
	}

	return nil
}

// CanToggleDHCP returns nil if you can toggle DHCP, or an error if you can't
func CanToggleDHCP() *nerr.E {
	otherFile := fmt.Sprintf("%s.other", dhcpFile)

	if _, err := os.Stat(dhcpFile); os.IsNotExist(err) {
		return nerr.Translate(err).Addf("can't toggle dhcp because there is no %s file", dhcpFile)
	}
	if _, err := os.Stat(otherFile); os.IsNotExist(err) {
		return nerr.Translate(err).Addf("can't toggle dhcp because there is no %s.other file", dhcpFile)
	}

	return nil
}
