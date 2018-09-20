package pi

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/structs"
)

var (
	deviceID    = os.Getenv("PI_HOSTNAME")
	roomIDregex = regexp.MustCompile(`.*-`)
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

// RoomID returns the room ID of the pi based on the hostname (everything before the last '-')
func RoomID() (string, *nerr.E) {
	id, err := DeviceID()
	if err != nil {
		return "", err.Addf("failed to get roomID")
	}

	reg := roomIDregex.Copy()
	matched := reg.FindAllString(id, -1)
	if len(matched) != 1 {
		return "", nerr.Create(fmt.Sprintf("something is wrong with the deviceID %s. My roomIDregex matched: %v", id, matched), "string")
	}

	return strings.TrimSuffix(matched[0], "-"), nil
}

// MustRoomID returns the buildingID or panics if it fails
func MustRoomID() string {
	id, err := RoomID()
	if err != nil {
		log.L.Fatalf("failed to get roomID: %s", err.Error())
	}

	return id
}
