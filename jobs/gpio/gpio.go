package gpio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"text/template"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/v2/events"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

// DividerSensorJob .
type DividerSensorJob struct{}

// Pin .
type Pin struct {
	Pin int `json:"pin"`

	Displays string `json:"displays"`
	Presets  struct {
		Connected    map[string]string `json:"connected"`
		Disconnected map[string]string `json:"disconnected"`
	} `json:"presets"`

	ReadFrequency     string `json:"read-frequency"`
	ReadsBeforeChange int    `json:"reads-before-change"`
	TrueUpFrequency   string `json:"true-up-frequency"`

	ChangeRequests []request `json:"change"`
	TrueUpRequests []request `json:"true-up"`

	Connected bool `json:"connected"`
}

type request struct {
	Method string      `json:"method"`
	URL    string      `json:"url"`
	Body   interface{} `json:"body"`
}

var (
	once    sync.Once
	pinList []*Pin
)

// Run .
func (j *DividerSensorJob) Run(ctx interface{}, eventWrite chan events.Event) interface{} {
	if ctx != nil {
		data, err := json.Marshal(ctx)
		if err != nil {
			return nerr.Translate(err).Addf("failed to run divider sensor job")
		}

		var pins []Pin
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
				pinList = append(pinList, &pins[i])
				go pins[i].monitor(adaptor)
			}
		})
	}

	return pinList
}

func (p *Pin) monitor(adaptor *raspi.Adaptor) {
	pin := gpio.NewDirectPinDriver(adaptor, strconv.Itoa(p.Pin))

	// get read duration
	readDuration, err := time.ParseDuration(p.ReadFrequency)
	if err != nil {
		log.L.Warnf("invalid read frequency: '%s'. defaulting to 200ms", p.ReadFrequency)
		readDuration = 200 * time.Millisecond // default read duration
	}
	log.L.Infof("Monitoring pin %v every %v", p.Pin, readDuration.String())

	// get true up duration
	trueUpDuration, err := time.ParseDuration(p.ReadFrequency)
	if err != nil {
		log.L.Warnf("invalid true up frequency: '%s'. defaulting to 5m", p.ReadFrequency)
		trueUpDuration = 5 * time.Minute // default true up duration
	}

	newStateCount := 0
	readTick := time.NewTicker(readDuration)
	trueUpTick := time.NewTicker(trueUpDuration)

	for {
		select {
		case <-readTick.C:
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

					for i := range p.ChangeRequests {
						go p.ChangeRequests[i].execute(p)
					}
				}
			}
		case <-trueUpTick.C:
			for i := range p.TrueUpRequests {
				go p.TrueUpRequests[i].execute(p)
			}
		}
	}
}

func (r *request) execute(pin *Pin) {
	url, err := pin.fillTemplate(r.URL)
	if err != nil {
		log.L.Warnf("failed to execute request against %s: %s", r.URL, err)
		return
	}

	log.L.Infof("Building %s request against %s", r.Method, url)
	body := ""

	// fill template if body is not nil
	if r.Body != nil {
		// turn body into json
		bytes, err := json.Marshal(r.Body)
		if err != nil {
			log.L.Warnf("unable to execute request against %s: %s", url, err)
			return
		}

		body, err = pin.fillTemplate(string(bytes))
		if err != nil {
			log.L.Warnf("unable to execute request against %s: %s", url, err)
			return
		}

		log.L.Debugf("Request body for %s: %s", url, body)
	}

	// build request
	req, err := http.NewRequest(r.Method, r.URL, bytes.NewReader([]byte(body)))
	if err != nil {
		log.L.Warnf("unable to execute request against %s: %s", url, err)
		return
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	log.L.Infof("Successfully built request. Sending %s request to %s", r.Method, url)
	resp, err := client.Do(req)
	if err != nil {
		log.L.Warnf("unable to execute request against %s: %s", url, err)
		return
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.L.Warnf("unable to read response from %s: %s", url, err)
		return
	}

	log.L.Infof("response from %s: %s", url, bytes)
}

func (p *Pin) fillTemplate(source string) (string, error) {
	t, err := template.New("pin").Parse(source)
	if err != nil {
		return "", fmt.Errorf("unable to fill template: %s", err)
	}

	buf := &bytes.Buffer{}
	err = t.Execute(buf, *p)
	if err != nil {
		return "", fmt.Errorf("unable to fill template: %s", err)
	}

	return buf.String(), nil
}

/* functions for to use in templates */

// Time returns the current time
func (p Pin) Time() string {
	return time.Now().Format(time.RFC3339)
}

// CurrentPreset returns the preset the hostname should be set to
func (p Pin) CurrentPreset(hostname string) string {
	if p.Connected {
		if v, ok := p.Presets.Connected[hostname]; ok {
			return v
		}
	} else {
		if v, ok := p.Presets.Disconnected[hostname]; ok {
			return v
		}
	}

	return "not a valid hostname"
}
