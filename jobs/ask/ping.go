package ask

import (
	"strings"
	"time"

	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/device-monitoring-microservice/pi"
	ping "github.com/sparrc/go-ping"
)

// PingJob is a job that pings each of the devices in the room and reports whether or not they are connected.
type PingJob struct {
	Count    int
	Interval time.Duration
	Timeout  time.Duration
}

// Run runs the job.
func (p *PingJob) Run(ctx interface{}, eventWrite chan events.Event) {
	devices, err := db.GetDB().GetDevicesByRoom(pi.MustRoomID())
	if err != nil {
		log.L.Warnf("error getting devices in room %v: %v", pi.MustRoomID(), err)
	}

	for _, device := range devices {
		if len(device.Address) == 0 || strings.EqualFold(device.Address, "0.0.0.0") {
			continue
		}

		pinger, err := ping.NewPinger(device.Address)
		if err != nil {
			log.L.Warnf("unable to create pinger for device %v (address: %v): %v", device.ID, device.Address, err)
			continue
		}

		pinger.Count = p.Count
		pinger.Interval = p.Interval
		pinger.Timeout = p.Timeout

		go pingTest(pinger, device, eventWrite)
	}
}

func pingTest(pinger *ping.Pinger, device structs.Device, eventWrite chan events.Event) {
	pinger.Run()

	event := events.Event{
		GeneratingSystem: pi.MustHostname(),
		Timestamp:        time.Now(),
		EventTags: []string{
			events.Heartbeat,
		},
		TargetRoom:   pi.MustRoomID(),
		TargetDevice: events.GenerateBasicDeviceInfo(pi.MustDeviceID()),
	}

	stats := pinger.Statistics()
	if stats.PacketLoss == 100.00 {
		log.L.Warnf("100%% packet loss to device %v (address %v)", device.ID, device.Address)
	} else {
		event.Key = "last-heartbeat"
		event.Value = time.Now().Format(time.RFC3339)
		eventWrite <- event
	}
}
