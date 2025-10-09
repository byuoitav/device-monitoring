package gpio

import (
	"strings"
	"time"

	"github.com/byuoitav/device-monitoring/localsystem"
)

func (p Pin) Time() string {
	return time.Now().Format(time.RFC3339)
}

func (p Pin) SystemID() string {
	return localsystem.MustSystemID()
}

func (p Pin) RoomID() string {
	return localsystem.MustRoomID()
}

func (p Pin) Room() string {
	id := localsystem.MustRoomID()
	split := strings.Split(id, "-")
	if len(split) == 2 {
		return split[1]
	}
	return id
}

func (p Pin) BuildingID() string {
	return localsystem.MustBuildingID()
}
