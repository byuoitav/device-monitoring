package ask

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/status"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/device-monitoring/pi"
)

// StatusJob checks the status of important microservices, and reports their status.
type StatusJob struct {
}

type statusConfig struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

// Run runs the job.
func (m *StatusJob) Run(ctx interface{}, eventWrite chan events.Event) interface{} {
	log.L.Infof("Getting status info...")

	microservices, ok := ctx.([]interface{})
	if !ok {
		return nerr.Create(fmt.Sprintf("bad context passed into status job: %v", ctx), "")
	}

	resultChan := make(chan status.Status, len(microservices))

	wg := &sync.WaitGroup{}
	for _, microservice := range microservices {
		data, ok := microservice.(map[string]interface{})
		if !ok {
			log.L.Warnf("one of the values in the context array is malformed.")
			continue
		}

		n, ok := data["name"].(string)
		if !ok || len(n) == 0 {
			log.L.Warnf("name must be a string with len > 0")
			continue
		}

		p := data["port"].(float64)
		if !ok || p == 0 {
			log.L.Warnf("port must be a float64 and not zero")
			continue
		}

		wg.Add(1)
		go func(name string, port float64) {
			// default values in case something goes wrong
			s := status.Status{
				Name:       name,
				StatusCode: status.Dead,
				Version:    `¯\_(ツ)_/¯`,
				Info:       make(map[string]interface{}),
			}
			log.L.Debugf("Getting %v status from port %v", name, port)

			defer func() {
				resultChan <- s
				wg.Done()
			}()

			// make request
			resp, err := http.Get(fmt.Sprintf("http://localhost:%v/status", port))
			if err != nil {
				s.Info["error"] = err.Error()
				return
			}

			// because we got a response, change to status sick
			s.StatusCode = status.Sick

			// read response
			bytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				s.Info["error"] = err.Error()
				return
			}
			resp.Body.Close()

			s.Info["_raw-response"] = fmt.Sprintf("%s", bytes)

			// unmarshal into status struct
			err = json.Unmarshal(bytes, &s)
			if err != nil {
				s.Info["error"] = err.Error()
				return
			}

			delete(s.Info, "_raw-response")

			// add back on the name
			s.Name = name
		}(n, p)
	}

	wg.Wait()
	close(resultChan)

	event := events.Event{
		GeneratingSystem: pi.MustHostname(),
		Timestamp:        time.Now(),
		EventTags: []string{
			events.Heartbeat,
		},
		AffectedRoom: events.GenerateBasicRoomInfo(pi.MustRoomID()),
		TargetDevice: events.GenerateBasicDeviceInfo(pi.MustDeviceID()),
	}

	var ret []status.Status
	for result := range resultChan {
		ret = append(ret, result)

		event.Key = fmt.Sprintf("%v-status", result.Name)
		event.Value = result.StatusCode
		event.Data = result
		eventWrite <- event
	}

	log.L.Infof("Finished getting status info.")

	return ret
}
