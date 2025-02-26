package gpio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"text/template"
	"time"

	"github.com/rs/zerolog/log"
)

// Pin represents a GPIO pin configuration.
type Pin struct {
	Pin  int  `json:"pin"`
	Flip bool `json:"flip"`

	BlueberryPresets string `json:"blueberry_presets"`
	Presets          struct {
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
	once sync.Once
	pins []Pin
)

// Monitor starts monitoring the signal on a pin
func (p *Pin) Monitor() {
	gpio := NewGPIO(p.Pin)
	gpio.Export()
	gpio.SetDirection(IN)

	// get read duration
	readDuration, err := time.ParseDuration(p.ReadFrequency)
	if err != nil {
		log.Warn().Msgf("Invalid read frequency: '%s'. Defaulting to 200ms", p.ReadFrequency)
		readDuration = 200 * time.Millisecond
	}
	log.Info().Msgf("Monitoring pin %v every %v", p.Pin, readDuration.String())

	// get true up duration
	trueUpDuration, err := time.ParseDuration(p.TrueUpFrequency)
	if err != nil {
		log.Warn().Msgf("Invalid true up frequency: '%s'. Defaulting to 5m", p.TrueUpFrequency)
		trueUpDuration = 5 * time.Minute
	}

	newStateCount := 0
	readTick := time.NewTicker(readDuration)
	trueUpTick := time.NewTicker(trueUpDuration)

	for {
		select {
		case <-readTick.C:
			read, err := gpio.Read()
			if err != nil {
				log.Warn().Msgf("Unable to read pin %v: %s", p.Pin, err)
				continue
			}

			connected := read == 1
			if p.Flip {
				connected = !connected
			}

			if connected != p.Connected {
				newStateCount++
				if newStateCount >= p.ReadsBeforeChange {
					newStateCount = 0
					p.Connected = connected
					log.Info().Msgf("Changed state to %v", p.Connected)

					for i := range p.ChangeRequests {
						go p.ChangeRequests[i].execute(p)
					}
				}
			}

		case <-trueUpTick.C:
			log.Info().Msg("Sending true-up requests.")
			for i := range p.TrueUpRequests {
				go p.TrueUpRequests[i].execute(p)
			}
		}
	}
}

func (r *request) execute(pin *Pin) {
	url, err := pin.fillTemplate(r.URL)
	if err != nil {
		log.Warn().Msgf("Failed to execute request against %s: %s", r.URL, err)
		return
	}

	log.Debug().Msgf("Building %s request against %s", r.Method, url)
	body := ""

	if r.Body != nil {
		bytes, err := json.Marshal(r.Body)
		if err != nil {
			log.Warn().Msgf("Unable to execute request against %s: %s", url, err)
			return
		}
		body = string(bytes)
	}

	req, err := http.NewRequest(r.Method, url, bytes.NewReader([]byte(body)))
	if err != nil {
		log.Warn().Msgf("Unable to execute request against %s: %s", url, err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	log.Info().Msgf("Sending %s request to %s", r.Method, url)
	resp, err := client.Do(req)
	if err != nil {
		log.Warn().Msgf("Unable to execute request against %s: %s", url, err)
		return
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Warn().Msgf("Unable to read response from %s: %s", url, err)
		return
	}

	log.Info().Msgf("Response from %s: %s", url, bytes)
}

func (p *Pin) fillTemplate(source string) (string, error) {
	t, err := template.New("pin").Parse(source)
	if err != nil {
		return "", fmt.Errorf("Unable to fill template: %s", err)
	}

	buf := &bytes.Buffer{}
	err = t.Execute(buf, *p)
	if err != nil {
		return "", fmt.Errorf("Unable to fill template: %s", err)
	}

	return buf.String(), nil
}

// SetPins sets the list of pins
func SetPins(p []Pin) {
	pins = p
}

// GetPins gets the list of pins
func GetPins() []Pin {
	return pins
}

func (p Pin) CurrentPreset(hostname string) string {
	for i := range p.ChangeRequests {
		if p.ChangeRequests[i].URL == hostname {
			return p.ChangeRequests[i].Method
		}
	}

	return ""
}
