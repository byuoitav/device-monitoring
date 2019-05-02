package ask

/*
// StatusJob checks the status of important microservices, and reports their status.
type StatusJob struct{}

type statusConfig struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

// Run runs the job.
func (m *StatusJob) Run(ctx interface{}, eventWrite chan events.Event) interface{} {
	event := events.Event{
		GeneratingSystem: localsystem.MustHostname(),
		Timestamp:        time.Now(),
		EventTags: []string{
			events.Heartbeat,
			events.Mstatus,
		},
		AffectedRoom: events.GenerateBasicRoomInfo(localsystem.MustRoomID()),
		TargetDevice: events.GenerateBasicDeviceInfo(localsystem.MustSystemID()),
	}

	var ret []status.Status
	for result := range resultChan {
		ret = append(ret, result)

		event.Key = fmt.Sprintf("%v-status", result.Name)
		event.Value = result.StatusCode
		event.Data = result
		eventWrite <- event
	}

	return ret
}
*/
