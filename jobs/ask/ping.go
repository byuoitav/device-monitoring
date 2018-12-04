package ask

import (
	"fmt"
	"math"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/device-monitoring/localsystem"
	ping "github.com/sparrc/go-ping"
)

// PingJob is a job that pings each of the devices in the room and reports whether or not they are connected.
type PingJob struct {
	Count    int
	Interval time.Duration
	Timeout  time.Duration
}

// PingResult is returned after running the ping job for a room
type PingResult struct {
	Successful   []devicePingResult `json:"successful,omitempty"`
	Unsuccessful []devicePingResult `json:"unsuccessful,omitempty"`
}

type devicePingResult struct {
	DeviceID string `json:"deviceID"`
	Error    string `json:"error"`

	PacketsReceived  int     `json:"packets-received"`
	PacketsSent      int     `json:"packets-sent"`
	PacketLoss       float64 `json:"packet-loss"`
	IPAddr           net.IP  `json:"ip"`
	Addr             string  `json:"address"`
	AverageRoundTrip string  `json:"average-round-trip"`
}

// Run runs the job.
func (p *PingJob) Run(ctx interface{}, eventWrite chan events.Event) interface{} {
	log.L.Infof("Getting pinggable status of devices in room %s", localsystem.MustRoomID())

	devices, err := db.GetDB().GetDevicesByRoom(localsystem.MustRoomID())
	if err != nil {
		return nerr.Translate(err).Addf("unable to get devices in room %v: %v", localsystem.MustRoomID(), err)
	}

	ret := PingResult{}
	resultChan := make(chan devicePingResult, len(devices))

	wg := &sync.WaitGroup{}
	for i := range devices {
		if len(devices[i].Address) == 0 || strings.EqualFold(devices[i].Address, "0.0.0.0") {
			continue
		}
		wg.Add(1)

		go func(index int) {
			defer wg.Done()

			result := devicePingResult{
				DeviceID: devices[index].ID,
			}

			// build pinger
			pinger, err := ping.NewPinger(devices[index].Address)
			if err != nil {
				result.Error = fmt.Sprintf("unable to create pinger for device %v (address: %v): %v", devices[index].ID, devices[index].Address, err)
				resultChan <- result
				return
			}

			pinger.Count = p.Count
			pinger.Interval = p.Interval
			pinger.Timeout = p.Timeout
			pinger.SetPrivileged(true)

			// run ping test
			pinger.Run()

			// parse results
			result.PacketsReceived = pinger.Statistics().PacketsRecv
			result.PacketsSent = pinger.Statistics().PacketsSent
			result.PacketLoss = pinger.Statistics().PacketLoss
			result.IPAddr = pinger.Statistics().IPAddr.IP
			result.Addr = pinger.Statistics().Addr
			result.AverageRoundTrip = fmt.Sprintf("%dms", pinger.Statistics().AvgRtt/time.Millisecond)

			// add in errors
			if math.IsNaN(result.PacketLoss) {
				result.PacketLoss = -1
				result.Error = fmt.Sprintf("unknown error, but packet loss was NaN")
			} else if result.PacketLoss == 100 {
				result.Error = fmt.Sprintf("no packets were successful")
			}

			resultChan <- result
		}(i)
	}

	wg.Wait()
	close(resultChan)

	// build a generic event to send for every device
	event := events.Event{
		GeneratingSystem: localsystem.MustSystemID(),
		Timestamp:        time.Now(),
		EventTags: []string{
			events.Heartbeat,
		},
		AffectedRoom: events.GenerateBasicRoomInfo(localsystem.MustRoomID()),
	}

	for result := range resultChan {
		if len(result.Error) > 0 {
			ret.Unsuccessful = append(ret.Unsuccessful, result)
			log.L.Infof("error pinging %v: %v", result.DeviceID, result.Error)
		} else {
			ret.Successful = append(ret.Successful, result)

			event.TargetDevice = events.GenerateBasicDeviceInfo(result.DeviceID)
			event.Key = "last-heartbeat"
			event.Value = time.Now().Format(time.RFC3339)
			eventWrite <- event
		}
	}

	// send an event for my own heartbeat
	event.TargetDevice = events.GenerateBasicDeviceInfo(localsystem.MustSystemID())
	event.Key = "last-heartbeat"
	event.Value = time.Now().Format(time.RFC3339)
	eventWrite <- event

	return ret
}
