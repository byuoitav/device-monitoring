package ask

/*

// StateUpdateJob is a job that gets the status of all the devices in the room, and pushes events from the status to "true up" the state of the room
type StateUpdateJob struct{}

// Run runs the job.
func (s *StateUpdateJob) Run(ctx interface{}, eventWrite chan events.Event) interface{} {
		// build base event
		event := events.Event{
			GeneratingSystem: localsystem.MustHostname(),
			Timestamp:        time.Now(),
			EventTags: []string{
				events.CoreState,
				events.AutoGenerated,
			},
			AffectedRoom: events.GenerateBasicRoomInfo(localsystem.MustRoomID()),
		}

		sentDisplays := make(map[string]bool)

		// take the response and send events
		log.L.Debugf("Sending events for displays...")
		for _, dev := range status.Displays {
			if strings.Contains(dev.Name, "-") {
				event.TargetDevice = events.GenerateBasicDeviceInfo(dev.Name)
			} else {
				event.TargetDevice = events.GenerateBasicDeviceInfo(fmt.Sprintf("%v-%v", localsystem.MustRoomID(), dev.Name))
			}

			log.L.Infof("Reporting status of %v", event.TargetDevice.DeviceID)

			// report power status
			if len(dev.Power) > 0 {
				event.Key = "power"
				event.Value = dev.Power
				eventWrite <- event
			}

			// report input status
			if len(dev.Input) > 0 {
				event.Key = "input"
				event.Value = dev.Input
				eventWrite <- event
			}

			// report blanked status
			if dev.Blanked != nil {
				event.Key = "blanked"
				event.Value = fmt.Sprintf("%v", *dev.Blanked)
				eventWrite <- event
			}

			sentDisplays[dev.Name] = true
		}

		log.L.Debugf("Sending events for audio devices...")
		for _, dev := range status.AudioDevices {
			if strings.Contains(dev.Name, "-") {
				event.TargetDevice = events.GenerateBasicDeviceInfo(dev.Name)
			} else {
				event.TargetDevice = events.GenerateBasicDeviceInfo(fmt.Sprintf("%v-%v", localsystem.MustRoomID(), dev.Name))
			}

			log.L.Infof("Reporting status of %v", event.TargetDevice.DeviceID)

			if dev.Muted != nil {
				event.Key = "muted"
				event.Value = fmt.Sprintf("%v", *dev.Muted)
				eventWrite <- event
			}

			if dev.Volume != nil {
				event.Key = "volume"
				event.Value = fmt.Sprintf("%v", *dev.Volume)
				eventWrite <- event
			}

			// send common info if it hasn't already been sent
			if _, ok := sentDisplays[dev.Name]; !ok {
				if len(dev.Power) > 0 {
					event.Key = "power"
					event.Value = dev.Power
					eventWrite <- event
				}

				if len(dev.Input) > 0 {
					event.Key = "input"
					event.Value = dev.Input
					eventWrite <- event
				}
			}
		}

		return status
	return nil
}
*/
