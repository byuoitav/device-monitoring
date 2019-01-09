package gpio

import (
	"sync"
	"time"

	"github.com/byuoitav/common/events"
	"github.com/byuoitav/common/log"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

// DividerSensorJob .
type DividerSensorJob struct{}

type sensorStatus struct {
	TrueStatus int
}

var (
	once *sync.Once
)

// Run .
func (j *DividerSensorJob) Run(ctx interface{}, eventWrite chan events.Event) interface{} {
	if ctx == nil {
		// just return the status
	}

	once.Do(func() {
		log.L.Infof("Monitoring divider sensors")
	})

	// return the status of the divider sensors

	return nil
}

// Read .
func Read() {
	log.SetLevel("info")
	r := raspi.NewAdaptor()
	sensor := gpio.NewDirectPinDriver(r, "7")

	for {
		time.Sleep(1 * time.Second)
		read, err := sensor.DigitalRead()
		if err != nil {
			log.L.Warnf("failed to read: %s", err)
			continue
		}

		log.L.Infof("state: %v", read)
	}
}
