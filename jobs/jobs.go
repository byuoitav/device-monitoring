package jobs

import (
	"sync"
	"time"

	"github.com/byuoitav/common/events"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/device-monitoring-microservice/jobs/ask"
)

// Job . :)
type Job interface {
	Run(ctx interface{})
}

var (
	jobs = map[string]Job{
		"ping": &ask.PingJob{
			Count:    4,
			Interval: 3 * time.Second,
			Timeout:  20 * time.Second,
		},
	}

	eventChan chan events.Event
)

// StartJobScheduler starts the jobs in the job map
func StartJobScheduler() {
	workers := 10
	queue := 100

	log.L.Infof("Starting job scheduler. Running %v jobs with %v workers with a max of %v events queued at once.", len(jobs), workers, queue)
	wg := sync.WaitGroup{}

	//	var matchJobs []Job
	eventChan = make(chan events.Event, queue)

	for _, job := range jobs {
		job.Run(nil)
	}

	// start event processors
	for i := 0; i < workers; i++ {
		log.L.Debugf("Starting event processor %v", i)
		wg.Add(1)

		/*
			go func() {
				defer wg.Done()

				for {
					select {
					case _ = <-eventChan:
						for _ := range matchJobs {
						}
					}
				}
			}()
		*/
	}

	wg.Wait()
}

// ProcessEvent passes the <event> in to be processed by the job scheduler
func ProcessEvent(event events.Event) {
	eventChan <- event
}
