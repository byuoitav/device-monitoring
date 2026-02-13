package gpio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// Monitor starts monitoring the signal on a pin
func (p *Pin) Monitor() {
	gpio := NewGPIO(p.Pin)
	if err := gpio.OpenInput(); err != nil {
		log.Warn().Err(err).Msgf("open gpio line %d", p.Pin)
	}
	defer gpio.Close()

	rd, err := time.ParseDuration(p.ReadFrequency)
	if err != nil {
		rd = 200 * time.Millisecond
	}
	td, err := time.ParseDuration(p.TrueUpFrequency)
	if err != nil {
		td = 5 * time.Minute
	}

	// Initial read
	if val, err := gpio.Read(); err == nil {
		connected := val == 1
		if p.Flip {
			connected = !connected
		}
		mu.Lock()
		p.Connected = connected
		mu.Unlock()
	}

	newStateCount := 0
	readTick := time.NewTicker(rd)
	trueUpTick := time.NewTicker(td)

	// Monitor loop
	for {
		select {
		case <-readTick.C:
			val, err := gpio.Read()
			if err != nil {
				log.Warn().Err(err).Msgf("read pin %d", p.Pin)
				continue
			}

			connected := val == 1
			if p.Flip {
				connected = !connected
			}

			mu.RLock()
			cur := p.Connected
			mu.RUnlock()

			if connected != cur {
				newStateCount++
				if newStateCount >= p.ReadsBeforeChange {
					newStateCount = 0
					mu.Lock()
					p.Connected = connected
					mu.Unlock()

					log.Info().Msgf("pin %d state changed -> %v", p.Pin, connected)
					for i := range p.ChangeRequests {
						go p.ChangeRequests[i].execute(p)
					}
				}
			}
			// NOTE: parity with old behavior — we do NOT reset newStateCount when equal

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
		log.Warn().Msgf("Failed to execute request against %s: %s", r.URL, err)
		return
	}

	log.Debug().Msgf("Building %s request against %s", r.Method, url)
	body := ""

	if r.Body != nil {
		// Marshal to JSON first, then allow {{.Field}} and pin methods in the JSON string.
		raw, err := json.Marshal(r.Body)
		if err != nil {
			log.Warn().Msgf("Unable to execute request against %s: %s", url, err)
			return
		}
		body, err = pin.fillTemplate(string(raw))
		if err != nil {
			log.Warn().Msgf("Unable to execute request against %s: %s", url, err)
			return
		}
		log.Debug().Msgf("Request body for %s: %s", url, body)
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

	reply, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Warn().Msgf("Unable to read response from %s: %s", url, err)
		return
	}
	log.Info().Msgf("Response from %s: %s", url, reply)
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
