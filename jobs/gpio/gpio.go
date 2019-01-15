package gpio

import (
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/v2/events"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

// DividerSensorJob .
type DividerSensorJob struct{}

type pin struct {
	Pin                int       `json:"pin"`
	Displays           string    `json:"displays"`
	ConnectRequests    []request `json:"connect-requests"`
	DisconnectRequests []request `json:"disconnect-requests"`

	ReadFrequency     string `json:"read-frequency"`
	ReadsBeforeChange int    `json:"reads-before-change"`

	Connected bool `json:"connected"`
}

type request struct {
	Method string      `json:"method"`
	URL    string      `json:"url"`
	Body   interface{} `json:"body"`
}

var (
	once sync.Once
)

// Run .
func (j *DividerSensorJob) Run(ctx interface{}, eventWrite chan events.Event) interface{} {
	if ctx != nil {
		data, err := json.Marshal(ctx)
		if err != nil {
			return nerr.Translate(err).Addf("failed to run divider sensor job")
		}

		var pins []pin
		err = json.Unmarshal(data, &pins)
		if err != nil {
			return nerr.Translate(err).Addf("failed to run divider sensor job")
		}

		if len(pins) == 0 {
			return nerr.Createf("empty", "failed to run divider sensor job - no pins configured")
		}

		once.Do(func() {
			log.L.Infof("Monitoring divider sensors")

			adaptor := raspi.NewAdaptor()

			for i := range pins {
				go pins[i].monitor(adaptor)
			}
		})
	}

	// return the status of the divider sensors
	return nil
}

func (p *pin) monitor(adaptor *raspi.Adaptor) {
	pin := gpio.NewDirectPinDriver(adaptor, strconv.Itoa(p.Pin))

	// get read duration
	duration, err := time.ParseDuration(p.ReadFrequency)
	if err != nil {
		log.L.Warnf("invalid read frequency: '%s'. defaulting to 200ms", p.ReadFrequency)
		duration = 200 * time.Millisecond // default duration
	}
	log.L.Infof("Monitoring pin %v every %v", p.Pin, duration.String())

	newStateCount := 0
	ticker := time.NewTicker(duration)

	for {
		select {
		case <-ticker.C:
			read, err := pin.DigitalRead()
			if err != nil {
				log.L.Warnf("unable to read pin %v: %s", p.Pin, err)
				continue
			}

			connected := read == 1

			// if the status is different than we thought it was
			if connected != p.Connected {
				newStateCount++

				if newStateCount >= p.ReadsBeforeChange {
					newStateCount = 0

					p.Connected = connected
					log.L.Infof("changed state to %v", p.Connected)

					if p.Connected {
						for i := range p.ConnectRequests {
							p.ConnectRequests[i].execute()
						}
					} else {
						for i := range p.DisconnectRequests {
							p.DisconnectRequests[i].execute()
						}
					}
				}
			}
		}
	}
}

func (r *request) execute() {
	log.SetLevel("info")
	log.L.Infof("executing %s request against %s", r.Method, r.Body)
	log.SetLevel("warn")
}
