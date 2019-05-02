package activesignal

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/inputgraph"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/device-monitoring/actions/roomstate"
	"github.com/byuoitav/device-monitoring/localsystem"
)

// TODO handle gated devices

const (
	activeSignalCommandID = "ActiveSignal"
)

// GetMap .
func GetMap(ctx context.Context) (map[string]bool, *nerr.E) {
	log.L.Infof("Getting active signal map")

	roomID, err := localsystem.RoomID()
	if err != nil {
		return nil, err.Addf("failed to get active signal info")
	}

	devices, gerr := db.GetDB().GetDevicesByRoom(roomID)
	if gerr != nil {
		return nil, nerr.Translate(gerr).Addf("failed to get active signal info")
	}

	graph, gerr := inputgraph.BuildGraph(devices, "video")
	if gerr != nil {
		return nil, nerr.Translate(gerr).Addf("failed to get active signal info")
	}

	// get current state of room
	state, err := roomstate.Get(ctx, roomID)
	if err != nil {
		return nil, err.Addf("failed to get active signal info")
	}

	activeMu := sync.Mutex{}
	active := make(map[string]bool)
	wg := sync.WaitGroup{}

	log.L.Infof("Got room state and build input graph, checking each display for active signal")

	for i := range state.Displays {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			deviceID := fmt.Sprintf("%s-%s", roomID, state.Displays[idx].Name)
			a := isInputPathActive(ctx, state.Displays[idx], roomID, graph)

			activeMu.Lock()
			active[deviceID] = a
			activeMu.Unlock()
		}(i)
	}

	wg.Wait()
	return active, nil
}

func isInputPathActive(ctx context.Context, display base.Display, roomID string, graph inputgraph.InputGraph) bool {
	if len(display.Input) == 0 || len(display.Name) == 0 {
		log.L.Debugf("Skipping %s because input or name is empty", display.Name)
		return false
	}

	displayID := fmt.Sprintf("%s-%s", roomID, display.Name)
	inputID := fmt.Sprintf("%s-%s", roomID, display.Input)

	log.L.Infof("Checking for active input from %s to %s", inputID, displayID)

	if display.Power == "standby" {
		log.L.Debugf("Input not active on %s because the power is %v", display.Name, display.Power)
		return false
	}

	if display.Blanked == nil || *display.Blanked {
		log.L.Debugf("Input not active on %s because blanked is true (or nil)", display.Name)
		return false
	}

	reachable, nodes, gerr := inputgraph.CheckReachability(displayID, inputID, graph)
	if gerr != nil {
		log.L.Warnf("failed to get active input information in room %s: %s", roomID, gerr)
		return false
	}

	if !reachable {
		log.L.Warnf("input is set to %s on display %s, but there is no valid path from that input to this display", inputID, displayID)
		return false
	}

	// loop from display to input
	for i := len(nodes) - 1; i >= 0; i-- {
		var src *structs.Device

		// TODO this needs to be more complex for just add power (does it?), etc
		if i != 0 {
			src = &nodes[i-1].Device
		}

		if !isInputActive(ctx, src, &nodes[i].Device) {
			log.L.Infof("There *is not* an active input signal from %s to %s", inputID, displayID)
			return false
		}
	}

	log.L.Infof("There *is* an active input signal from %s to %s", inputID, displayID)
	return true
}

// isInputActive returns true if the port connecting dest -> src is marked as active
// if src is nil, then it returns true if dest claims there is an active input
func isInputActive(ctx context.Context, src *structs.Device, dest *structs.Device) bool {
	if dest == nil {
		log.L.Errorf("destination device passed into isInputActive cannot be null.")
		return false
	}

	// TODO maybe cache whether or not a specific input as active for a little while?
	// create a new sub logger
	l := log.L.Named(dest.ID)

	if src != nil {
		l.Debugf("Checking if %s is sending an active input signal to me", src.ID)
	} else {
		l.Debugf("Checking if I'm sending an active input signal")
	}

	address := dest.GetCommandByID(activeSignalCommandID).BuildCommandAddress()
	if len(address) == 0 {
		// for now, if the command doesn't exist, we are going to assume the input was active
		l.Debugf("I do not have a %s command. Let's assume I'm sending a signal", activeSignalCommandID)
		return true
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

	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		l.Warnf("unable to check if input was active: %s", err)
		return false
	}

	req = req.WithContext(ctx)
	c := http.Client{
		Timeout: 5 * time.Second,
	}

	l.Debugf("Sending GET request to: %s", address)
	resp, err := c.Do(req)
	if err != nil {
		l.Warnf("unable to check if input was active: %s", err)
		return false
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Warnf("unable to check if input was active: %s", err)
		return false
	}

	var active structs.ActiveSignal
	err = json.Unmarshal(bytes, &active)
	if err != nil {
		l.Warnf("unable to check if input was active: %s. response body: %s", err, bytes)
		return false
	}

	if src != nil && active.Active {
		l.Debugf("%s *is* sending an active input signal to me", src.ID)
	} else if src == nil && active.Active {
		l.Debugf("I *am* sending an active input signal")
	} else if src != nil {
		l.Debugf("%s *is not* sending an active input signal to me", src.ID)
	} else {
		l.Debugf("I *am not* sending an active input signal")
	}

	return active.Active
}
