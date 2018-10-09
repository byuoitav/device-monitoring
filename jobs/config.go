package jobs

import (
	"time"

	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/device-monitoring-microservice/jobs/ask"
)

// Job . :)
type Job interface {
	Run(ctx interface{}, eventWrite chan events.Event)
}

var (
	jobs = map[string]Job{
		"ping": &ask.PingJob{
			Count:    4,
			Interval: 3 * time.Second,
			Timeout:  20 * time.Second,
		},
		"state-update": &ask.StateUpdateJob{},
		"mstatus":      &ask.MStatusJob{},
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
	Type  string      `json:"type"`  // required for all
	At    string      `json:"at"`    // required for 'time'
	Every string      `json:"every"` // required for 'interval'
	Match MatchConfig `json:"match"` // required for 'event'
}
