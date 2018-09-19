package ask

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/status"
	"github.com/byuoitav/common/v2/events"
)

// MStatusJob checks the mstatus of important microservices, and reports their status.
type MStatusJob struct {
}

var microserviceURLs = map[string]int{
	"av-api":            8000,
	"event-router":      7000,
	"touchpanel-ui":     8888,
	"event-translator":  6998,
	"device-monitoring": 10000,
}

// Run runs the job.
func (m *MStatusJob) Run(ctx interface{}, eventWrite chan events.Event) {
	log.L.Infof("Getting mstatus info...")

	hostname, _ := os.Hostname()
	event := events.Event{
		GeneratingSystem: hostname,
		Timestamp:        time.Now(),
		EventTags: []string{
			events.Heartbeat,
		},
		TargetRoom: events.BasicRoomInfo{
			BuildingID: buildingID,
			RoomID:     roomID,
		},
		TargetDevice: events.GenerateBasicDeviceInfo(hostname),
	}

	for name, port := range microserviceURLs {
		event.Key = fmt.Sprintf("%s-mstatus", name)
		event.Value = status.Dead // default status

		// make request
		resp, err := http.Get(fmt.Sprintf("http://localhost:%v/mstatus", port))
		if err != nil {
			event.Data = err
			eventWrite <- event
			continue
		}

		// because we got a response, change to status sick
		event.Value = status.Sick

		// read response
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			event.Data = err
			eventWrite <- event
			continue
		}
		resp.Body.Close()

		// unmarshal into status struct
		var s status.MStatus
		err = json.Unmarshal(bytes, &s)
		if err != nil {
			event.Data = err
			eventWrite <- event
			continue
		}

		// send up real status
		event.Value = s.StatusCode
		event.Data = s
		eventWrite <- event
	}
}
