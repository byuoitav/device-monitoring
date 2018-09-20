package pi

import (
	"os"

	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/structs"
)

// PiHostname returns the pi hostname of the device
func PiHostname() (string, *nerr.E) {
	pihn := os.Getenv("PI_HOSTNAME")
	if len(pihn) == 0 {
		return "", nerr.Create("PI_HOSTNAME not set.", "string")
	}

	vals := structs.DeviceIDValidationRegex.FindStringSubmatch(pihn)
	if len(vals) == 0 {
		return "", nerr.Create("PI_HOSTNAME is set as %s, which is an invalid hostname.", pihn)
	}

	return pihn, nil
}
