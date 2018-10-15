package ask

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/status"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/device-monitoring/pi"
)

// MStatusJob checks the mstatus of important microservices, and reports their status.
type MStatusJob struct {
}

type mStatusConfig struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

// Run runs the job.
func (m *MStatusJob) Run(ctx interface{}, eventWrite chan events.Event) interface{} {
	log.L.Infof("Getting mstatus info...")

	microservices, ok := ctx.([]interface{})
	if !ok {
		return nerr.Create(fmt.Sprintf("bad context passed into mstatus job: %v", ctx), "")
	}

	event := events.Event{
		GeneratingSystem: pi.MustHostname(),
		Timestamp:        time.Now(),
		EventTags: []string{
			events.Heartbeat,
		},
		AffectedRoom: events.GenerateBasicRoomInfo(pi.MustRoomID()),
		TargetDevice: events.GenerateBasicDeviceInfo(pi.MustDeviceID()),
	}

	for _, microservice := range microservices {
		data, ok := microservice.(map[string]interface{})
		if !ok {
			log.L.Warnf("One of the values in the context array is malformed.")
			continue
		}

		name := data["name"]
		port := data["port"]

		event.Key = fmt.Sprintf("%v-mstatus", name)
		event.Value = status.Dead // default status

		log.L.Debugf("Getting %v, using port %v", event.Key, port)

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

	log.L.Infof("Finished getting mstatus info.")

	return nil
}
