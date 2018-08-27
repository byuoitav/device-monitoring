package jobs

import (
	"fmt"
	"regexp"

	"github.com/byuoitav/common/events"
)

// MatchConfig contains the logic for building/matching regex for events that come in
type MatchConfig struct {
	Count int

	Hostname         string `json:"hostname,omitempty"`
	Timestamp        string `json:"timestamp,omitempty"`
	LocalEnvironment string `json:"localEnvironment,omitempty"`
	Building         string `json:"building,omitempty"`
	Room             string `json:"room,omitempty"`

	Event struct {
		Type           string `json:"type,omitempty"`
		Requestor      string `json:"requestor,omitempty"`
		EventCause     string `json:"eventCause,omitempty"`
		Device         string `json:"device,omitempty"`
		EventInfoKey   string `json:"eventInfoKey,omitempty"`
		EventInfoValue string `json:"eventInfoValue,omitempty"`
	} `json:"event,omitempty"`

	Regex struct {
		Hostname         *regexp.Regexp
		Timestamp        *regexp.Regexp
		LocalEnvironment *regexp.Regexp
		Building         *regexp.Regexp
		Room             *regexp.Regexp

		Event struct {
			Type           *regexp.Regexp
			Requestor      *regexp.Regexp
			EventCause     *regexp.Regexp
			Device         *regexp.Regexp
			EventInfoKey   *regexp.Regexp
			EventInfoValue *regexp.Regexp
		}
	}
}

func (r *runner) buildMatchRegex() {
	r.Trigger.Match.Count = 0

	// build the regex for each field
	if len(r.Trigger.Match.Hostname) > 0 {
		r.Trigger.Match.Regex.Hostname = regexp.MustCompile(r.Trigger.Match.Hostname)
		r.Trigger.Match.Count++
	}

	if len(r.Trigger.Match.Timestamp) > 0 {
		r.Trigger.Match.Regex.Timestamp = regexp.MustCompile(r.Trigger.Match.Timestamp)
		r.Trigger.Match.Count++
	}

	if len(r.Trigger.Match.LocalEnvironment) > 0 {
		r.Trigger.Match.Regex.LocalEnvironment = regexp.MustCompile(r.Trigger.Match.LocalEnvironment)
		r.Trigger.Match.Count++
	}

	if len(r.Trigger.Match.Building) > 0 {
		r.Trigger.Match.Regex.Building = regexp.MustCompile(r.Trigger.Match.Building)
		r.Trigger.Match.Count++
	}

	if len(r.Trigger.Match.Room) > 0 {
		r.Trigger.Match.Regex.Room = regexp.MustCompile(r.Trigger.Match.Room)
		r.Trigger.Match.Count++
	}

	if len(r.Trigger.Match.Event.Type) > 0 {
		r.Trigger.Match.Regex.Event.Type = regexp.MustCompile(r.Trigger.Match.Event.Type)
		r.Trigger.Match.Count++
	}

	if len(r.Trigger.Match.Event.Requestor) > 0 {
		r.Trigger.Match.Regex.Event.Requestor = regexp.MustCompile(r.Trigger.Match.Event.Requestor)
		r.Trigger.Match.Count++
	}

	if len(r.Trigger.Match.Event.EventCause) > 0 {
		r.Trigger.Match.Regex.Event.EventCause = regexp.MustCompile(r.Trigger.Match.Event.EventCause)
		r.Trigger.Match.Count++
	}

	if len(r.Trigger.Match.Event.Device) > 0 {
		r.Trigger.Match.Regex.Event.Device = regexp.MustCompile(r.Trigger.Match.Event.Device)
		r.Trigger.Match.Count++
	}

	if len(r.Trigger.Match.Event.EventInfoKey) > 0 {
		r.Trigger.Match.Regex.Event.EventInfoKey = regexp.MustCompile(r.Trigger.Match.Event.EventInfoKey)
		r.Trigger.Match.Count++
	}

	if len(r.Trigger.Match.Event.EventInfoValue) > 0 {
		r.Trigger.Match.Regex.Event.EventInfoValue = regexp.MustCompile(r.Trigger.Match.Event.EventInfoValue)
		r.Trigger.Match.Count++
	}
}

func (r *runner) doesEventMatch(event *events.Event) bool {
	if r.Trigger.Match.Count == 0 {
		return true
	}

	if r.Trigger.Match.Regex.Hostname != nil {
		reg := r.Trigger.Match.Regex.Hostname.Copy()
		if !reg.MatchString(event.Hostname) {
			return false
		}
	}

	if r.Trigger.Match.Regex.Timestamp != nil {
		reg := r.Trigger.Match.Regex.Timestamp.Copy()
		if !reg.MatchString(event.Timestamp) {
			return false
		}
	}

	if r.Trigger.Match.Regex.LocalEnvironment != nil {
		reg := r.Trigger.Match.Regex.LocalEnvironment.Copy()
		if !reg.MatchString(fmt.Sprintf("%v", event.LocalEnvironment)) {
			return false
		}
	}

	if r.Trigger.Match.Regex.Building != nil {
		reg := r.Trigger.Match.Regex.Building.Copy()
		if !reg.MatchString(event.Building) {
			return false
		}
	}

	if r.Trigger.Match.Regex.Room != nil {
		reg := r.Trigger.Match.Regex.Room.Copy()
		if !reg.MatchString(event.Room) {
			return false
		}
	}

	if r.Trigger.Match.Regex.Event.Type != nil {
		reg := r.Trigger.Match.Regex.Event.Type.Copy()
		if !reg.MatchString(fmt.Sprintf("%v", event.Event.Type)) {
			return false
		}
	}

	if r.Trigger.Match.Regex.Event.Requestor != nil {
		reg := r.Trigger.Match.Regex.Event.Requestor.Copy()
		if !reg.MatchString(event.Event.Requestor) {
			return false
		}
	}

	if r.Trigger.Match.Regex.Event.EventCause != nil {
		reg := r.Trigger.Match.Regex.Event.EventCause.Copy()
		if !reg.MatchString(fmt.Sprintf("%v", event.Event.EventCause)) {
			return false
		}
	}

	if r.Trigger.Match.Regex.Event.Device != nil {
		reg := r.Trigger.Match.Regex.Event.Device.Copy()
		if !reg.MatchString(event.Event.Device) {
			return false
		}
	}

	if r.Trigger.Match.Regex.Event.EventInfoKey != nil {
		reg := r.Trigger.Match.Regex.Event.EventInfoKey.Copy()
		if !reg.MatchString(event.Event.EventInfoKey) {
			return false
		}
	}

	if r.Trigger.Match.Regex.Event.EventInfoValue != nil {
		reg := r.Trigger.Match.Regex.Event.EventInfoValue.Copy()
		if !reg.MatchString(event.Event.EventInfoValue) {
			return false
		}
	}

	return true
}
