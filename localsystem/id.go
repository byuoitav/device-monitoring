package localsystem

import (
	"fmt"
	"os"
	"strings"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/structs"
)

const (
	// EnvironmentFile is the location of the environment file for the av-control api
	EnvironmentFile = "/avcapi/environment"
)

var (
	systemID    = os.Getenv("SYSTEM_ID")
	installerID = os.Getenv("INSTALLER_ID")
)

// SystemID returns the pi hostname of the device
func SystemID() (string, *nerr.E) {
	if len(systemID) == 0 {
		return "", nerr.Create("SYSTEM_ID not set.", "string")
	}

	if !structs.IsDeviceIDValid(systemID) {
		return "", nerr.Create("SYSTEM_ID is set as %s, which is an invalid hostname.", systemID)
	}

	return systemID, nil
}

// MustSystemID returns the pi hostname of the device, and panics if it isn't set or is invalid
func MustSystemID() string {
	id, err := SystemID()
	if err != nil {
		log.L.Fatalf("%s", err.Error())
	}

	return id
}

// BuildingID returns the room ID of the pi based on the hostname (everything before the last '-')
func BuildingID() (string, *nerr.E) {
	id, err := SystemID()
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
	id, err := SystemID()
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

// InstallerID returns the installerID of the pi
func InstallerID() (string, *nerr.E) {
	if len(installerID) == 0 {
		return "", nerr.Create("INSTALLER_ID not set.", "string")
	}

	return installerID, nil
}

// MustInstallerID returns the installerID or the pi or panics if it fails
func MustInstallerID() string {
	id, err := InstallerID()
	if err != nil {
		log.L.Fatalf("failed to get installerID: %s", err.Error())
	}

	return id
}

// SetInstallerID sets the installer id
func SetInstallerID(id string) *nerr.E {
	errMsg := "failed to set installer ID"

	file, err := os.OpenFile(EnvironmentFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nerr.Translate(err).Add(errMsg)
	}

	// TODO do i need to delete an old installer id, or should i just leave it on
	_, err = file.WriteString(fmt.Sprintf(`INSTALLER_ID="%s"`, id))
	if err != nil {
		return nerr.Translate(err).Add(errMsg)
	}
	file.Close()

	err = os.Setenv("INSTALLER_ID", id)
	if err != nil {
		return nerr.Translate(err).Add(errMsg)
	}

	return nil
}
