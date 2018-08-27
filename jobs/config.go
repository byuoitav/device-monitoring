package jobs

import (
	"time"

	"github.com/byuoitav/common/events"
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
		"monitoring": &ask.MonitoringJob{},
	}
)

// JobConfig defines a configuration of a specific job.
type JobConfig struct {
	Name     string    `json:"name"`
	Triggers []Trigger `json:"triggers"`
	Enabled  bool      `json:"enabled"`
}

// Trigger matches something that causes a job to be ran.
type Trigger struct {
	Type  string      `json:"type"`  // required for all
	At    string      `json:"at"`    // required for 'time'
	Every string      `json:"every"` // required for 'interval'
	Match MatchConfig `json:"match"` // required for 'event'
}
