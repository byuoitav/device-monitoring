package pi

import (
	"os"
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
