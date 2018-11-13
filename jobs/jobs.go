package jobs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/central-event-system/hub/base"
	"github.com/byuoitav/central-event-system/messenger"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/v2/events"
)

var (
	runners []*runner
	configs []JobConfig

	m *messenger.Messenger
)

type runner struct {
	Job          Job
	Config       JobConfig
	Trigger      Trigger
	TriggerIndex int
}

func init() {
	// TODO check if there is a config in couchdb first
	// parse configuration
	path := os.Getenv("JOB_CONFIG_LOCATION")
	if len(path) < 1 {
		path = "./config.json" // default config location
	}
	log.L.Infof("Parsing job configuration from %v", path)

	// read configuration
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.L.Fatalf("failed to read job configuration: %v", err)
	}

	// unmarshal job config
	err = json.Unmarshal(b, &configs)
	if err != nil {
		log.L.Fatalf("unable to parse job configuration: %v", err)
	}

	// validate all jobs exist
	for _, config := range configs {
		if !config.Enabled {
			log.L.Debugf("Skipping %v, because it's disabled.", config.Name)
			continue
		}

		if _, ok := jobs[config.Name]; !ok {
			log.L.Fatalf("job %v doesn't exist.", config.Name)
		}

		// build a runner for each trigger
		for i, trigger := range config.Triggers {
			runner := &runner{
				Job:          jobs[config.Name],
				Config:       config,
				Trigger:      trigger,
				TriggerIndex: i,
			}

			// build regex if it's a match type
			if strings.EqualFold(runner.Trigger.Type, "match") {
				runner.buildMatchRegex()
			}

			log.L.Infof("Adding runner for job '%v', trigger #%v. Execution Type: %v", runner.Config.Name, runner.TriggerIndex, runner.Trigger.Type)
			runners = append(runners, runner)
		}
	}
}

// StartJobScheduler starts the jobs in the job map
func StartJobScheduler() {
	// start event router
	hubAddr := os.Getenv("HUB_ADDRESS")
	if len(hubAddr) == 0 {
		log.L.Fatalf("HUB_ADDRESS is not set.")
	}

	var err *nerr.E
	m, err = messenger.BuildMessenger(hubAddr, base.Messenger, 1000)
	if err != nil {
		fmt.Printf("Could not build the messenger: %s", err)
	}

	workers := 10
	queue := 100

	log.L.Infof("Starting job scheduler. Running %v jobs with %v workers with a max of %v events queued at once.", len(jobs), workers, queue)
	wg := sync.WaitGroup{}

	var matchRunners []*runner
	for _, runner := range runners {
		switch runner.Trigger.Type {
		case "daily":
			go runner.runDaily()
		case "interval":
			go runner.runInterval()
		case "match":
			matchRunners = append(matchRunners, runner)
		default:
			log.L.Warnf("unknown trigger type '%v' for job %v|%v", runner.Trigger.Type, runner.Config.Name, runner.TriggerIndex)
		}
	}

	eventChan := make(chan events.Event, 300)
	go readEvents(eventChan)

	// start event processors
	for i := 0; i < workers; i++ {
		log.L.Debugf("Starting event processor %v", i)
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				select {
				case event := <-eventChan:
					for i := range matchRunners {
						if matchRunners[i].doesEventMatch(&event) {
							go matchRunners[i].run(&event)
						}
					}
				}
			}
		}()
	}

	wg.Wait()
}

// RunJob takes a job and runs it with the given context
func RunJob(job Job, ctx interface{}) interface{} {
	eventChan := make(chan events.Event, 100)
	defer close(eventChan)

	go func() {
		for event := range eventChan {
			log.L.Debugf("Publishing event: %+v", event)
			m.SendEvent(event)
		}
	}()

	return job.Run(ctx, eventChan)
}

// TODO this needs to be updated with new event package
func readEvents(outChan chan events.Event) {
	for {
		event := m.ReceiveEvent()
		outChan <- event
	}
}

func (r *runner) run(context interface{}) {
	log.L.Debugf("[%s|%v] Running job...", r.Config.Name, r.TriggerIndex)

	resp := RunJob(r.Job, context)
	switch v := resp.(type) {
	case error:
		log.L.Warnf("failed to run job: %s", v)
	case *nerr.E:
		log.L.Warnf("failed to run job: %s", v.String())
	case nerr.E:
		log.L.Warnf("failed to run job: %s", v.String())
	}

	log.L.Debugf("[%s|%v] Finished.", r.Config.Name, r.TriggerIndex)
}

func (r *runner) runDaily() {
	tmpDate := fmt.Sprintf("2006-01-02T%s", r.Trigger.At)
	runTime, err := time.Parse(time.RFC3339, tmpDate)
	runTime = runTime.UTC()
	if err != nil {
		log.L.Warnf("unable to parse time '%s' to execute job %s daily. error: %s", r.Trigger.At, r.Config.Name, err)
		return
	}

	log.L.Infof("[%s|%v] Running daily at %s", r.Config.Name, r.TriggerIndex, runTime.Format("15:04:05 MST"))

	// figure out how long until next run
	now := time.Now()
	until := time.Until(time.Date(now.Year(), now.Month(), now.Day(), runTime.Hour(), runTime.Minute(), runTime.Second(), 0, runTime.Location()))
	if until < 0 {
		until = 24*time.Hour + until
	}

	log.L.Debugf("[%s|%v] Time to next run: %v", r.Config.Name, r.TriggerIndex, until)
	timer := time.NewTimer(until)

	for {
		<-timer.C
		r.run(r.Config.Context)

		timer.Reset(24 * time.Hour)
	}
}

func (r *runner) runInterval() {
	interval, err := time.ParseDuration(r.Trigger.Every)
	if err != nil {
		log.L.Warnf("unable to parse duration '%s' to execute job %s on an interval. error: %s", r.Trigger.Every, r.Config.Name, err)
		return
	}

	log.L.Infof("[%s|%v] Running every %v", r.Config.Name, r.TriggerIndex, interval)

	ticker := time.NewTicker(interval)
	for range ticker.C {
		r.run(r.Config.Context)
	}
}

// EventNode returns the event node.
func Messenger() *messenger.Messenger {
	return m
}

// GetJobContext returns the context parsed for a specific job, even if it isn't enabled
func GetJobContext(job string) interface{} {
	for i := range configs {
		if strings.EqualFold(configs[i].Name, job) {
			return configs[i].Context
		}
	}

	return nil
}
