package ask

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
	ping "github.com/sparrc/go-ping"
)

var room string

func init() {
	hostname := os.Getenv("PI_HOSTNAME")
	if len(hostname) == 0 {
		log.L.Fatalf("PI_HOSTNAME not set.")
	}

	split := strings.Split(hostname, "-")
	if len(split) != 3 {
		log.L.Fatalf("hostname %v is formed incorrectly. Should match BLDG-ROOM-DEVICE.", hostname)
	}

	room = fmt.Sprintf("%v-%v", split[0], split[1])
}

// PingJob is a job that pings each of the devices in the room and reports whether or not they are connected.
type PingJob struct {
	Count    int
	Interval time.Duration
	Timeout  time.Duration
}

// Run runs the job.
func (p *PingJob) Run(ctx interface{}) {
	devices, err := db.GetDB().GetDevicesByRoom(room)
	if err != nil {
		log.L.Warnf("error getting devices in room %v: %v", room, err)
	}

	for _, device := range devices {
		if len(device.Address) == 0 || strings.EqualFold(device.Address, "0.0.0.0") {
			continue
		}

		pinger, err := ping.NewPinger(device.Address)
		if err != nil {
			log.L.Warnf("unable to create pinger for device %v (address: %v): %v", device.ID, device.Address, err)
		}

		pinger.Count = p.Count
		pinger.Interval = p.Interval
		pinger.Timeout = p.Timeout

		go pingTest(pinger, device)
	}
}

func pingTest(pinger *ping.Pinger, device structs.Device) {
	pinger.Run()

	stats := pinger.Statistics()
	if stats.PacketLoss == 100.00 {
		log.L.Warnf("100%% packet loss to device %v (address %v)", device.ID, device.Address)
	}
}
