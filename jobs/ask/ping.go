package ask

import (
	"context"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/device-monitoring/actions/then/ping"
	"github.com/byuoitav/device-monitoring/localsystem"
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

var (
	permCheck sync.Once
	root      bool
)

// Run runs the job.
func (p *PingJob) Run(ctx interface{}, eventWrite chan events.Event) interface{} {
	// check permissions
	permCheck.Do(func() {
		root = os.Getuid() == 0
	})

	if !root {
		return nerr.Createf("error", "insufficient permissions to ping; please run program as root user")
	}

	log.L.Infof("Getting pinggable status of devices in room %s", localsystem.MustRoomID())
	defer log.L.Infof("Finished ping job.")

	devices, err := db.GetDB().GetDevicesByRoom(localsystem.MustRoomID())
	if err != nil {
		return nerr.Translate(err).Addf("unable to get devices in room %v: %v", localsystem.MustRoomID(), err)
	}

	// ret := PingResult{}
	// resultChan := make(chan devicePingResult, len(devices))

	hosts := []string{}
	for i := range devices {
		if len(devices[i].Address) == 0 || strings.EqualFold(devices[i].Address, "0.0.0.0") {
			continue
		}
		hosts = append(hosts, devices[i].Address)
	}

	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pinger, err := ping.NewPinger()
	if err != nil {
		log.L.Fatal(err)
	}

	results := pinger.Ping(c, hosts...)
	return results
}
