package ask

import (
	"fmt"
	"net/http"
	"os"
	"time"

	events2 "github.com/byuoitav/common/v2/events"
)

// MStatusJob checks the mstatus of important microservices, and reports their status.
type MStatusJob struct {
}

var microserviceURLs = map[string]string{
	"av-api":        "localhost:8000",
	"event-router":  "localhost:7000",
	"touchpanel-ui": "localhost:8888",
}

// Run runs the job.
func (m *MStatusJob) Run(ctx interface{}, eventWrite chan events2.Event) {
	hostname, _ := os.Hostname()

	event := events2.Event{
		GeneratingSystem: hostname,
		Timestamp:        time.Now(),
		EventTags: []string{
			events2.Heartbeat,
		},
		TargetRoom: events2.BasicRoomInfo{
			BuildingID: buildingID,
			RoomID:     roomID,
		},
		TargetDevice: events2.GenerateBasicDeviceInfo(hostname),
	}

	for _, url := range microserviceURLs {
		resp, err := http.Get(fmt.Sprintf("http://%v", url))
		if err != nil {
		}
	}
}
