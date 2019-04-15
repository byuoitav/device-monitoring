package ask

import (
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
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/device-monitoring/localsystem"
)

const (
	activeSignalCommandID = "ActiveSignal"
)

// ActiveSignalJob asks each device what it's current input status to decide if the current input for each device is correct
type ActiveSignalJob struct{}

// Run runs the job
func (j *ActiveSignalJob) Run(ctx interface{}, eventWrite chan events.Event) interface{} {
	log.L.Infof("Starting active input job")
	systemID, err := localsystem.SystemID()
	if err != nil {
		return err.Addf("failed to get active input information")
	}

	roomID, err := localsystem.RoomID()
	if err != nil {
		return err.Addf("failed to get active input information")
	}

	devices, gerr := db.GetDB().GetDevicesByRoom(roomID)
	if gerr != nil {
		return nerr.Translate(gerr).Addf("failed to get active input information in room %s", roomID)
	}

	graph, gerr := inputgraph.BuildGraph(devices, "video")
	if err != nil {
		return nerr.Translate(gerr).Addf("failed to get active input inforamtion in room %s", roomID)
	}

	// get current status of room
	roomState := base.PublicRoom{}
	stateJob := &StateUpdateJob{}
	state := stateJob.Run(ctx, eventWrite)

	switch v := state.(type) {
	case error:
		return nerr.Translate(v).Addf("failed to get active input information in room %s", roomID)
	case *nerr.E:
		return v.Addf("failed to get active input information in room %s", roomID)
	case base.PublicRoom:
		roomState = v
	case *base.PublicRoom:
		roomState = *v
	default:
		nerr.Translate(fmt.Errorf("something went wrong getting current status of room %s: %s", roomID, v)).Addf("failed to get active input information in room %s", roomID)
	}

	wg := sync.WaitGroup{}
	activeMutex := sync.Mutex{}
	activeMap := make(map[string]bool)

	log.L.Infof("Got room state and built input graph, checking each display for active input")
	for i := range roomState.Displays {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			active := isInputPathActive(roomState.Displays[idx], roomID, graph)
			deviceID := fmt.Sprintf("%s-%s", roomID, roomState.Displays[idx].Name)

			activeMutex.Lock()
			activeMap[deviceID] = active
			activeMutex.Unlock()

			eventWrite <- events.Event{
				GeneratingSystem: systemID,
				Timestamp:        time.Now(),
				EventTags: []string{
					events.DetailState,
				},
				TargetDevice: events.GenerateBasicDeviceInfo(deviceID),
				AffectedRoom: events.GenerateBasicRoomInfo(roomID),
				Key:          "active-signal",
				Value:        fmt.Sprintf("%v", active),
			}
		}(i)
	}
	wg.Wait()

	return activeMap
}

func isInputPathActive(display base.Display, roomID string, graph inputgraph.InputGraph) bool {
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

		// TODO this needs to be more complex for jap (does it?), etc
		if i != 0 {
			src = &nodes[i-1].Device
		}

		if !isInputActive(src, &nodes[i].Device) {
			log.L.Infof("There *is not* an active input signal from %s to %s", inputID, displayID)
			return false
		}
	}

	log.L.Infof("There *is* an active input signal from %s to %s", inputID, displayID)
	return true
}

// isInputActive returns true if the port connecting dest -> src is marked as active
// if src is nil, then it returns true if dest claims there is an active input
func isInputActive(src *structs.Device, dest *structs.Device) bool {
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

	c := http.Client{
		Timeout: 20 * time.Second,
	}

	l.Debugf("Sending GET request to: %s", address)
	resp, err := c.Get(address)
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

func cachedIsInputActive() *bool {
	return nil
}

func cacheInputActive() {
}
