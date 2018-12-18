package ask

import (
	"github.com/byuoitav/av-api/inputgraph"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/device-monitoring/localsystem"
)

// TODO the question is: should we just push up each devices information? or should we only look at the devices that we care about

// ActiveInputJob asks each device what it's current input status to decide if the current input for each device is correct
type ActiveInputJob struct{}

// Run runs the job
func (j *ActiveInputJob) Run(ctx interface{}, eventWrite chan events.Event) interface{} {
	log.L.Infof("Starting active input job")
	roomID, err := localsystem.RoomID()
	if err != nil {
		return err.Addf("failed to get active input information")
	}

	devices, gerr := db.GetDB().GetDevicesByRoom(roomID)
	if gerr != nil {
		return nerr.Translate(gerr).Addf("failed to get active input information in room %s", roomID)
	}

	graph, gerr := inputgraph.BuildGraph(devices)
	if err != nil {
		return nerr.Translate(gerr).Addf("failed to get active input inforamtion in room %s", roomID)
	}

	log.L.Debugf("Input graph built:")
	for k, v := range graph.AdjacencyMap {
		log.L.Debugf("%s -> %s", k, v)
	}

	reachable, nodes, gerr := inputgraph.CheckReachability("ITB-1101-D1", "ITB-1101-VIA1", graph)
	if gerr != nil {
		return nerr.Translate(gerr).Addf("failed to get active input inforamtion in room %s", roomID)
	}

	if reachable {
		log.L.Infof("reachability path:")
		for i := len(nodes) - 1; i >= 0; i-- {
			log.L.Infof("%s", nodes[i].ID)
		}
	}

	// address := fmt.Sprintf("http://localhost:8000/buildings/%s/rooms/%s")
	return nil
}
