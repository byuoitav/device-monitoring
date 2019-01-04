package jobs

import (
	"time"

	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/device-monitoring/jobs/ask"
)

// Job . :)
type Job interface {
	Run(ctx interface{}, eventWrite chan events.Event) interface{}
}

var (
	jobs = map[string]Job{
		"ping": &ask.PingJob{
			Count:    4,
			Interval: 3 * time.Second,
			Timeout:  20 * time.Second,
		},
		"state-update":         &ask.StateUpdateJob{},
		"status":               &ask.StatusJob{},
		"device-hardware-info": &ask.DeviceHardwareJob{},
		"hardware-info":        &ask.HardwareInfoJob{},
		"active-signal":        &ask.ActiveSignalJob{},
	}
)

// JobConfig defines a configuration of a specific job.
type JobConfig struct {
	Name     string      `json:"name"`
	Triggers []Trigger   `json:"triggers"`
	Enabled  bool        `json:"enabled"`
	Context  interface{} `json:"context"`
}

// Trigger matches something that causes a job to be ran.
type Trigger struct {
	Type  string       `json:"type"`            // required for all
	At    *string      `json:"at,omitempty"`    // required for 'time'
	Every *string      `json:"every,omitempty"` // required for 'interval'
	Match *MatchConfig `json:"match,omitempty"` // required for 'event'
}

// RunnerInfo contains info about a specific runner
type RunnerInfo struct {
	ID      string      `json:"id"`
	Trigger Trigger     `json:"trigger"`
	Context interface{} `json:"context,omitempty"`

	RunnerStatus
}

// RunnerStatus is the status of a runner
type RunnerStatus struct {
	LastRunStartTime *time.Time `json:"last-run-start-time,omitempty"`
	LastRunDuration  string     `json:"last-run-duration"`
	LastRunError     string     `json:"last-run-error"`
	CurrentlyRunning bool       `json:"currently-running"`
	RunCount         int        `json:"run-count"`
}
