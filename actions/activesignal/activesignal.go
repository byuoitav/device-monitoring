package activesignal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/inputgraph"
	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/device-monitoring/actions/roomstate"
	"github.com/byuoitav/device-monitoring/couchdb"
	"github.com/byuoitav/device-monitoring/localsystem"
)

const (
	activeSignalCommandID = "ActiveSignal"
)

// GetMap .
func GetMap(ctx context.Context) (map[string]bool, error) {
	slog.Info("Getting active signal map")

	roomID, err := localsystem.RoomID()
	if err != nil {
		return nil, fmt.Errorf("failed to get active signal info: %w could not get room ID", err)
	}

	devices, gerr := couchdb.GetDevicesByRoom(ctx, roomID)
	if gerr != nil {
		return nil, fmt.Errorf("failed to get active signal info: %w could not get devices in room", gerr)
	}

	graph, gerr := inputgraph.BuildGraph(devices, "video")
	if gerr != nil {
		return nil, fmt.Errorf("failed to get active signal info: %w could not build input graph", gerr)
	}

	// get current state of room
	state, err := roomstate.Get(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active signal info: %w could not get room state", err)
	}

	activeMu := sync.Mutex{}
	active := make(map[string]bool)
	wg := sync.WaitGroup{}

	slog.Info("Got room state and build input graph, checking each display for active signal")

	for i := range state.Displays {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			deviceID := fmt.Sprintf("%s-%s", roomID, state.Displays[idx].Name)
			a := isInputPathActive(ctx, state.Displays[idx], roomID, graph, devices)

			activeMu.Lock()
			active[deviceID] = a
			activeMu.Unlock()
		}(i)
	}

	wg.Wait()
	return active, nil
}

func isInputPathActive(ctx context.Context, display base.Display, roomID string, graph inputgraph.InputGraph, roomDevices []structs.Device) bool {
	if len(display.Input) == 0 || len(display.Name) == 0 {
		slog.Debug("Skipping display because input or name is empty", slog.String("displayName", display.Name))
		return false
	}

	displayID := fmt.Sprintf("%s-%s", roomID, display.Name)
	inputID := fmt.Sprintf("%s-%s", roomID, display.Input)

	slog.Info("Checking for active input from %s to %s", inputID, displayID)

	if display.Power == "standby" {
		slog.Debug("Input not active because the power is standby", slog.String("displayName", display.Name))
		return false
	}

	if display.Blanked == nil || *display.Blanked {
		slog.Debug("Input not active because blanked is true (or nil)", slog.String("displayName", display.Name))
		return false
	}

	reachable, nodes, gerr := inputgraph.CheckReachability(displayID, inputID, graph)
	if gerr != nil {
		slog.Warn("failed to get active input information", slog.String("roomID", roomID), slog.String("error", gerr.Error()))
		return false
	}

	if !reachable {
		slog.Warn("input is not reachable from display", slog.String("displayID", displayID), slog.String("inputID", inputID))
		return false
	}

	// loop from display to input
	for i := len(nodes) - 1; i >= 0; i-- {
		var src *structs.Device

		// TODO this needs to be more complex for just add power (does it?), etc
		if i != 0 {
			src = &nodes[i-1].Device
		}

		if !isInputActive(ctx, src, &nodes[i].Device, roomDevices) {
			slog.Info("There *is not* an active input signal from %s to %s", inputID, displayID)
			return false
		}
	}

	slog.Info("There *is* an active input signal from %s to %s", inputID, displayID)
	return true
}

// isInputActive returns true if the port connecting dest -> src is marked as active
// if src is nil, then it returns true if dest claims there is an active input
func isInputActive(ctx context.Context, src *structs.Device, dest *structs.Device, roomDevices []structs.Device) bool {
	if dest == nil {
		slog.Error("destination device passed into isInputActive cannot be null.")
		return false
	}

	// TODO maybe cache whether or not a specific input as active for a little while?

	// create a new sub logger for this device
	l := slog.With(
		slog.String("destID", dest.ID),
	)

	if src != nil {
		l.Debug("Checking if source is sending an active input signal to me", slog.String("srcID", src.ID))
	} else {
		l.Debug("Checking if I am sending an active input signal")
	}

	if !dest.HasCommand(activeSignalCommandID) {
		return true // assume that the signal is active if we can't check it
	}

	address, err := dest.BuildCommandURL(activeSignalCommandID)
	if err != nil {
		l.Warn("unable to check if input is active", slog.String("error", err.Error()))
		return false
	}

	address = strings.Replace(address, ":address", dest.Address, 1)

	if src != nil && strings.Contains(address, ":port") {
		port := dest.GetPortFromSrc(src.ID)
		if port == nil {
			// shouldn't ever get here, we validated that there was a path/port earlier
			return false
		}

		portID := port.ID

		if structs.HasRole(*dest, "VideoSwitcher") {
			// Replace IN/OUT so we just pass the port number to videoswitchers
			portID = strings.Replace(portID, "IN", "", 1)
			portID = strings.Replace(portID, "OUT", "", 1)
		}

		address = strings.Replace(address, ":port", portID, 1)
	}

	req, reqErr := http.NewRequest("GET", address, nil)
	if reqErr != nil {
		l.Warn("unable to check if input was active", slog.String("error", reqErr.Error()))
		return false
	}

	req = req.WithContext(ctx)
	c := http.Client{
		Timeout: 5 * time.Second,
	}

	l.Debug("Sending GET request", slog.String("address", address))
	resp, respErr := c.Do(req)
	if respErr != nil {
		l.Warn("unable to check if input was active", slog.String("error", respErr.Error()))
		return false
	}
	defer resp.Body.Close()

	bytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		l.Warn("unable to check if input was active", slog.String("error", readErr.Error()))
		return false
	}

	var active structs.ActiveSignal
	unmarshalErr := json.Unmarshal(bytes, &active)
	if unmarshalErr != nil {
		l.Warn("unable to check if input was active", slog.String("error", unmarshalErr.Error()), slog.String("responseBody", string(bytes)))
		return false
	}

	if src != nil && active.Active {
		l.Debug(fmt.Sprintf("%s *is* sending an active input signal to me", src.ID), slog.String("srcID", src.ID))
	} else if src == nil && active.Active {
		l.Debug("I *am* sending an active input signal")
	} else if src != nil {
		l.Debug(fmt.Sprintf("%s *is not* sending an active input signal to me", src.ID), slog.String("srcID", src.ID))
	} else {
		l.Debug("I *am not* sending an active input signal")
	}

	return active.Active
}
