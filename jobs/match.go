package jobs

import (
	"encoding/json"
	"regexp"

	"github.com/byuoitav/common/v2/events"
)

// MatchConfig contains the logic for building/matching regex for events that come in
type MatchConfig struct {
	Count int

	GeneratingSystem string   `json:"generating-system"`
	Timestamp        string   `json:"timestamp"`
	EventTags        []string `json:"event-tags"`
	Key              string   `json:"key"`
	Value            string   `json:"value"`
	User             string   `json:"user"`
	Data             string   `json:"data,omitempty"`
	AffectedRoom     struct {
		BuildingID string `json:"buildingID,omitempty"`
		RoomID     string `json:"roomID,omitempty"`
	} `json:"affected-room"`
	TargetDevice struct {
		BuildingID string `json:"buildingID,omitempty"`
		RoomID     string `json:"roomID,omitempty"`
		DeviceID   string `json:"deviceID,omitempty"`
	} `json:"target-device"`

	Regex struct {
		GeneratingSystem *regexp.Regexp
		Timestamp        *regexp.Regexp
		EventTags        []*regexp.Regexp
		Key              *regexp.Regexp
		Value            *regexp.Regexp
		User             *regexp.Regexp
		Data             *regexp.Regexp
		AffectedRoom     struct {
			BuildingID *regexp.Regexp
			RoomID     *regexp.Regexp
		}
		TargetDevice struct {
			BuildingID *regexp.Regexp
			RoomID     *regexp.Regexp
			DeviceID   *regexp.Regexp
		}
	}
}

func (r *runner) buildMatchRegex() {
	r.Trigger.Match.Count = 0

	// build the regex for each field
	if len(r.Trigger.Match.GeneratingSystem) > 0 {
		r.Trigger.Match.Regex.GeneratingSystem = regexp.MustCompile(r.Trigger.Match.GeneratingSystem)
		r.Trigger.Match.Count++
	}

	if len(r.Trigger.Match.Timestamp) > 0 {
		r.Trigger.Match.Regex.Timestamp = regexp.MustCompile(r.Trigger.Match.Timestamp)
		r.Trigger.Match.Count++
	}

	if len(r.Trigger.Match.Key) > 0 {
		r.Trigger.Match.Regex.Key = regexp.MustCompile(r.Trigger.Match.Key)
		r.Trigger.Match.Count++
	}

	if len(r.Trigger.Match.Value) > 0 {
		r.Trigger.Match.Regex.Value = regexp.MustCompile(r.Trigger.Match.Value)
		r.Trigger.Match.Count++
	}

	if len(r.Trigger.Match.User) > 0 {
		r.Trigger.Match.Regex.User = regexp.MustCompile(r.Trigger.Match.User)
		r.Trigger.Match.Count++
	}

	if len(r.Trigger.Match.Data) > 0 {
		r.Trigger.Match.Regex.Data = regexp.MustCompile(r.Trigger.Match.Data)
		r.Trigger.Match.Count++
	}

	if len(r.Trigger.Match.TargetDevice.BuildingID) > 0 {
		r.Trigger.Match.Regex.TargetDevice.BuildingID = regexp.MustCompile(r.Trigger.Match.TargetDevice.BuildingID)
		r.Trigger.Match.Count++
	}

	if len(r.Trigger.Match.TargetDevice.RoomID) > 0 {
		r.Trigger.Match.Regex.TargetDevice.RoomID = regexp.MustCompile(r.Trigger.Match.TargetDevice.RoomID)
		r.Trigger.Match.Count++
	}

	if len(r.Trigger.Match.TargetDevice.DeviceID) > 0 {
		r.Trigger.Match.Regex.TargetDevice.DeviceID = regexp.MustCompile(r.Trigger.Match.TargetDevice.DeviceID)
		r.Trigger.Match.Count++
	}

	if len(r.Trigger.Match.AffectedRoom.BuildingID) > 0 {
		r.Trigger.Match.Regex.AffectedRoom.BuildingID = regexp.MustCompile(r.Trigger.Match.AffectedRoom.BuildingID)
		r.Trigger.Match.Count++
	}

	if len(r.Trigger.Match.AffectedRoom.RoomID) > 0 {
		r.Trigger.Match.Regex.AffectedRoom.RoomID = regexp.MustCompile(r.Trigger.Match.AffectedRoom.RoomID)
		r.Trigger.Match.Count++
	}

	for _, tag := range r.Trigger.Match.EventTags {
		r.Trigger.Match.Regex.EventTags = append(r.Trigger.Match.Regex.EventTags, regexp.MustCompile(tag))
		r.Trigger.Match.Count++
	}
}

func (r *runner) doesEventMatch(event *events.Event) bool {
	if r.Trigger.Match.Count == 0 {
		return true
	}

	if r.Trigger.Match.Regex.GeneratingSystem != nil {
		reg := r.Trigger.Match.Regex.GeneratingSystem.Copy()
		if !reg.MatchString(event.GeneratingSystem) {
			return false
		}
	}

	if r.Trigger.Match.Regex.Timestamp != nil {
		reg := r.Trigger.Match.Regex.Timestamp.Copy()
		if !reg.MatchString(event.Timestamp.String()) {
			return false
		}
	}

	if len(r.Trigger.Match.Regex.EventTags) > 0 {
		matched := 0

		for _, regex := range r.Trigger.Match.Regex.EventTags {
			reg := regex.Copy()

			for _, tag := range event.EventTags {
				if reg.MatchString(tag) {
					matched++
					break
				}
			}
		}

		if matched != len(r.Trigger.Match.Regex.EventTags) {
			return false
		}
	}

	if r.Trigger.Match.Regex.Key != nil {
		reg := r.Trigger.Match.Regex.Key.Copy()
		if !reg.MatchString(event.Key) {
			return false
		}
	}

	if r.Trigger.Match.Regex.Value != nil {
		reg := r.Trigger.Match.Regex.Value.Copy()
		if !reg.MatchString(event.Value) {
			return false
		}
	}

	if r.Trigger.Match.Regex.User != nil {
		reg := r.Trigger.Match.Regex.User.Copy()
		if !reg.MatchString(event.User) {
			return false
		}
	}

	if r.Trigger.Match.Regex.Data != nil {
		reg := r.Trigger.Match.Regex.Data.Copy()
		// convert event.Data to a json string
		bytes, err := json.Marshal(event.Data)
		if err != nil {
			return false
		}

		// or should i do fmt.Sprintf?
		if !reg.MatchString(string(bytes[:])) {
			return false
		}
	}

	if r.Trigger.Match.Regex.TargetDevice.BuildingID != nil {
		reg := r.Trigger.Match.Regex.TargetDevice.BuildingID.Copy()
		if !reg.MatchString(event.TargetDevice.BuildingID) {
			return false
		}
	}

	if r.Trigger.Match.Regex.TargetDevice.RoomID != nil {
		reg := r.Trigger.Match.Regex.TargetDevice.RoomID.Copy()
		if !reg.MatchString(event.TargetDevice.RoomID) {
			return false
		}
	}

	if r.Trigger.Match.Regex.TargetDevice.DeviceID != nil {
		reg := r.Trigger.Match.Regex.TargetDevice.DeviceID.Copy()
		if !reg.MatchString(event.TargetDevice.DeviceID) {
			return false
		}
	}

	if r.Trigger.Match.Regex.AffectedRoom.BuildingID != nil {
		reg := r.Trigger.Match.Regex.AffectedRoom.BuildingID.Copy()
		if !reg.MatchString(event.AffectedRoom.BuildingID) {
			return false
		}
	}

	if r.Trigger.Match.Regex.AffectedRoom.RoomID != nil {
		reg := r.Trigger.Match.Regex.AffectedRoom.RoomID.Copy()
		if !reg.MatchString(event.AffectedRoom.RoomID) {
			return false
		}
	}

	return true
}
