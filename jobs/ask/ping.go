package ask

import (
	"net"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/v2/events"
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

// Run runs the job.
func (p *PingJob) Run(ctx interface{}, eventWrite chan events.Event) interface{} {
	log.L.Infof("Getting pinggable status of devices in room %s", localsystem.MustRoomID())
	defer log.L.Infof("Finished ping job.")

	return nil
}
