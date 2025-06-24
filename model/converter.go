package model

// Converter converts messages from the central-event-system to the device-monitoring model.Event type.
// we are trying to move away from the common library, so this is a temporary solution
// until everything is moved over to their own model types.

import (
	"log/slog"

	"github.com/byuoitav/common/v2/events"
)

// ConvertEvent transforms an events.Event from the central-event-system into a device-monitoring model.Event.
// logs an error if a key is not found in the event.
func ConvertEvent(e events.Event) Event {
	if e.GeneratingSystem == "" {
		slog.Error("event has no generating system", slog.Any("event", e))
		return Event{}
	}
	if e.Timestamp.IsZero() {
		slog.Error("event has no timestamp", slog.Any("event", e))
		return Event{}
	}
	return Event{
		GeneratingSystem: e.GeneratingSystem,
		Timestamp:        e.Timestamp,
		EventTags:        append([]string{}, e.EventTags...),
		TargetDevice:     GenerateBasicDeviceInfo(e.TargetDevice.DeviceID),
		AffectedRoom:     GenerateBasicRoomInfo(e.AffectedRoom.RoomID),
		Key:              e.Key,
		Value:            e.Value,
		User:             e.User,
		Data:             e.Data,
	}
}

func ConvertEvents(events []events.Event) []Event {
	converted := make([]Event, len(events))
	for i, e := range events {
		converted[i] = ConvertEvent(e)
	}
	return converted
}

func ToCommonEvent(e Event) events.Event {
	// sanity checks
	if e.GeneratingSystem == "" {
		slog.Warn("Converting model.Event with empty GeneratingSystem", slog.Any("event", e))
	}
	if e.Key == "" {
		slog.Warn("Converting model.Event with empty Key", slog.Any("event", e))
	}

	return events.Event{
		GeneratingSystem: e.GeneratingSystem,
		Timestamp:        e.Timestamp,
		EventTags:        append([]string{}, e.EventTags...),
		TargetDevice:     events.GenerateBasicDeviceInfo(e.TargetDevice.DeviceID),
		AffectedRoom:     events.GenerateBasicRoomInfo(e.AffectedRoom.RoomID),
		Key:              e.Key,
		Value:            e.Value,
		User:             e.User,
		Data:             e.Data,
	}
}
