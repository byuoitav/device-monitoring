package model

// Converter converts messages from the central-event-system to the device-monitoring model.Event type.
// we are trying to move away from the common library, so this is a temporary solution
// until everything is moved over to their own model types.

import (
	"log/slog"

	"github.com/byuoitav/common/structs"
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

// ConvertDevice converts a structs.Device from the common library to a model.Device.
func ConvertDevice(d structs.Device) Device {
	cmds := make([]Command, len(d.Type.Commands))
	for i, c := range d.Type.Commands {
		cmds[i] = Command{ID: c.ID}
	}
	return Device{
		ID:      d.ID,
		Address: d.Address,
		Type:    DeviceType{Commands: cmds},
		Proxy:   d.Proxy,
	}
}

// ConvertDevices maps a slice of structs.Device to a slice of model.Device.
func ConvertDevices(devices []structs.Device) []Device {
	converted := make([]Device, len(devices))
	for i, d := range devices {
		converted[i] = ConvertDevice(d)
	}
	return converted
}

func ToCommonDevice(d Device) structs.Device {
	cmds := make([]structs.Command, len(d.Type.Commands))
	for i, c := range d.Type.Commands {
		cmds[i] = structs.Command{ID: c.ID}

	}
	return structs.Device{
		ID:      d.ID,
		Address: d.Address,
		Type: structs.DeviceType{
			Commands: cmds,
			// everything else stays at its zero‐value
		},
		Proxy: d.Proxy,
		// Name, Description, Roles, Ports, etc. will be empty
	}
}

func ToCommonDevices(devices []Device) []structs.Device {
	converted := make([]structs.Device, len(devices))
	for i, d := range devices {
		converted[i] = ToCommonDevice(d)
	}
	return converted
}
