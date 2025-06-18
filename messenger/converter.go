package messenger

// Converter converts messages from the central-event-system to the device-monitoring model.Event type.
// we are trying to move away from the common library, so this is a temporary solution
// until everything is moved over to their own model types.

import (
	"log/slog"

	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/device-monitoring/model"
)

// ConvertEvent transforms an events.Event from the central-event-system into a device-monitoring model.Event.
// logs an error if a key is not found in the event.
func ConvertEvent(e events.Event) model.Event {
	if e.GeneratingSystem == "" {
		slog.Error("event has no generating system", slog.Any("event", e))
		return model.Event{}
	}
	if e.Timestamp.IsZero() {
		slog.Error("event has no timestamp", slog.Any("event", e))
		return model.Event{}
	}
	return model.Event{
		GeneratingSystem: e.GeneratingSystem,
		Timestamp:        e.Timestamp,
		EventTags:        append([]string{}, e.EventTags...),
		TargetDevice:     model.GenerateBasicDeviceInfo(e.TargetDevice.DeviceID),
		AffectedRoom:     model.GenerateBasicRoomInfo(e.AffectedRoom.RoomID),
		Key:              e.Key,
		Value:            e.Value,
		User:             e.User,
		Data:             e.Data,
	}

}
