package localsystem

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/byuoitav/device-monitoring/model"
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
func SystemID() (string, error) {
	if systemID == "" {
		return "", fmt.Errorf("SYSTEM_ID not set")
	}
	if !model.IsDeviceIDValid(systemID) {
		return "", fmt.Errorf("SYSTEM_ID %q is not a valid device ID", systemID)
	}
	return systemID, nil
}

// MustSystemID returns the pi hostname of the device, and panics if it isn't set or is invalid
func MustSystemID() string {
	id, err := SystemID()
	if err != nil {
		log.Fatalf("failed to get system ID: %v", err)
	}

	return id
}

// BuildingID returns the room ID of the pi based on the hostname (everything before the last '-')
func BuildingID() (string, error) {
	id, err := SystemID()
	if err != nil {
		return "", fmt.Errorf("failed to get building ID: %w", err)
	}
	parts := strings.SplitN(id, "-", 2)
	return parts[0], nil
}

// MustBuildingID returns the buildingID or panics if it fails
func MustBuildingID() string {
	id, err := BuildingID()
	if err != nil {
		log.Fatalf("failed to get building ID: %v", err)
	}

	return id
}

// RoomID returns the room ID of the pi based on the hostname (everything before the last '-')
func RoomID() (string, error) {
	id, err := SystemID()
	if err != nil {
		return "", fmt.Errorf("failed to get room ID: %w", err)
	}
	parts := strings.SplitN(id, "-", 3)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid SYSTEM_ID format for room ID: %q", id)
	}
	return parts[0] + "-" + parts[1], nil
}

// MustRoomID returns the buildingID or panics if it fails
func MustRoomID() string {
	id, err := RoomID()
	if err != nil {
		log.Fatalf("failed to get room ID: %v", err)
	}

	return id
}

// InstallerID returns the installerID of the pi
func InstallerID() (string, error) {
	if installerID == "" {
		return "", fmt.Errorf("INSTALLER_ID not set")
	}
	return installerID, nil
}

// MustInstallerID returns the installerID or the pi or panics if it fails
func MustInstallerID() string {
	id, err := InstallerID()
	if err != nil {
		log.Fatalf("failed to get installer ID: %v", err)
	}

	return id
}

// SetInstallerID sets the installer id
func SetInstallerID(id string) error {
	f, err := os.OpenFile(EnvironmentFile, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", EnvironmentFile, err)
	}
	defer f.Close()

	line := fmt.Sprintf(`INSTALLER_ID="%s"`+"\n", id)
	if _, err := f.WriteString(line); err != nil {
		return fmt.Errorf("failed to write installer ID to %s: %w", EnvironmentFile, err)
	}

	if err := os.Setenv("INSTALLER_ID", id); err != nil {
		return fmt.Errorf("failed to set INSTALLER_ID env var: %w", err)
	}

	slog.Info("Set INSTALLER_ID", slog.String("installer_id", id))
	return nil
}
