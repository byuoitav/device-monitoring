package healthinfrastructure

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

func SendSuccessfulStartup(healthCheck func() map[string]string, MicroserviceName string, publish func(eventinfrastructure.Event)) error {
	log.Printf("[HealthCheck] will report success in 10 seconds, waiting for listening services to be up")
	time.Sleep(10 * time.Second)
	log.Printf("[HealthCheck] Reporting microsrevice startup complete")

	log.Printf("[HealthCheck] Checking Health...")
	statusReport := healthCheck()
	allSuccess := true
	for _, v := range statusReport {
		if v != "ok" {
			allSuccess = false
		}
	}

	report := make(map[string]interface{})
	if allSuccess {
		report["success"] = "ok"
	} else {
		report["success"] = "errors"
	}
	report["report"] = statusReport
	report["Microservice"] = MicroserviceName

	log.Printf("[HealthCheck] Reporting...")
	for k, v := range statusReport {
		publishEvent(publish, k, v, MicroserviceName)
	}

	if allSuccess {
		publishEvent(publish, "ready", "true", MicroserviceName)
	} else {
		publishEvent(publish, "ready", "false", MicroserviceName)
	}
	return nil
}

func publishEvent(publish func(eventinfrastructure.Event), k string, v string, name string) {
	publish(BuildEvent(
		eventinfrastructure.HEALTH,
		eventinfrastructure.STARTUP,
		k, v, name,
	))
}

func BuildEvent(Type eventinfrastructure.EventType, Cause eventinfrastructure.EventCause, Key string, Value string, Device string) eventinfrastructure.Event {

	info := eventinfrastructure.EventInfo{
		Type:           Type,
		EventCause:     Cause,
		Device:         Device,
		EventInfoKey:   Key,
		EventInfoValue: Value,
	}

	hostname := os.Getenv("PI_HOSTNAME")
	split := strings.Split(hostname, "-")

	return eventinfrastructure.Event{
		Hostname:         hostname,
		Timestamp:        time.Now().Format(time.RFC3339),
		LocalEnvironment: len(os.Getenv("LOCAL_ENVIRONMENT")) > 0,
		Event:            info,
		Building:         split[0],
		Room:             split[1],
	}

}
