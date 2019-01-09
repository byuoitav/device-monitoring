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
	log.SetLevel("info")
	if ctx != nil {
		data, err := json.Marshal(ctx)
		if err != nil {
			return nerr.Translate(err).Addf("failed to start divider sensor job")
		}

		var pins []pin
		err = json.Unmarshal(data, &pins)
		if err != nil {
			return nerr.Translate(err).Addf("failed to start divider sensor job")
		}

		if len(pins) == 0 {
			return nerr.Createf("empty", "failed to start divider sensor job - no pins configured")
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
	ticker := time.NewTicker(1 * time.Second)
	pin := gpio.NewDirectPinDriver(adaptor, strconv.Itoa(p.Pin))

	log.L.Infof("Monitoring pin %v", p.Pin)

	for {
		select {
		case <-ticker.C:
			read, err := pin.DigitalRead()
			if err != nil {
				log.L.Warnf("unable to read pin %v: %s", p.Pin, err)
				continue
			}

			log.L.Infof("read: %v", read)
		}
	}
}
