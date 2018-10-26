package pi

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/structs"
)

var (
	deviceID = os.Getenv("PI_HOSTNAME")
)

// DeviceID returns the pi hostname of the device
func DeviceID() (string, *nerr.E) {
	if len(deviceID) == 0 {
		return "", nerr.Create("PI_HOSTNAME not set.", "string")
	}

	if !structs.IsDeviceIDValid(deviceID) {
		return "", nerr.Create("PI_HOSTNAME is set as %s, which is an invalid hostname.", deviceID)
	}

	return deviceID, nil
}

// MustDeviceID returns the pi hostname of the device, and panics if it isn't set or is invalid
func MustDeviceID() string {
	id, err := DeviceID()
	if err != nil {
		log.L.Fatalf("%s", err.Error())
	}

	return id
}

// BuildingID returns the room ID of the pi based on the hostname (everything before the last '-')
func BuildingID() (string, *nerr.E) {
	id, err := DeviceID()
	if err != nil {
		return "", err.Addf("failed to get buildingID")
	}

	split := strings.Split(id, "-")
	return split[0], nil
}

// MustBuildingID returns the buildingID or panics if it fails
func MustBuildingID() string {
	id, err := BuildingID()
	if err != nil {
		log.L.Fatalf("failed to get buildingID: %s", err.Error())
	}

	return id
}

// RoomID returns the room ID of the pi based on the hostname (everything before the last '-')
func RoomID() (string, *nerr.E) {
	id, err := DeviceID()
	if err != nil {
		return "", err.Addf("failed to get roomID")
	}

	split := strings.Split(id, "-")
	return split[0] + "-" + split[1], nil
}

// MustRoomID returns the buildingID or panics if it fails
func MustRoomID() string {
	id, err := RoomID()
	if err != nil {
		log.L.Fatalf("failed to get roomID: %s", err.Error())
	}

	return id
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
