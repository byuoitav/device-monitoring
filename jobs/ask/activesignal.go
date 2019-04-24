package ask

/*

// ActiveSignalJob asks each device what it's current input status to decide if the current input for each device is correct
type ActiveSignalJob struct{}

// Run runs the job
func (j *ActiveSignalJob) Run(ctx interface{}, eventWrite chan events.Event) interface{} {
	for i := range roomState.Displays {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			active := isInputPathActive(roomState.Displays[idx], roomID, graph)
			deviceID := fmt.Sprintf("%s-%s", roomID, roomState.Displays[idx].Name)

			activeMutex.Lock()
			activeMap[deviceID] = active
			activeMutex.Unlock()

			eventWrite <- events.Event{
				GeneratingSystem: systemID,
				Timestamp:        time.Now(),
				EventTags: []string{
					events.DetailState,
				},
				TargetDevice: events.GenerateBasicDeviceInfo(deviceID),
				AffectedRoom: events.GenerateBasicRoomInfo(roomID),
				Key:          "active-signal",
				Value:        fmt.Sprintf("%v", active),
			}
		}(i)
	}
	wg.Wait()

	return activeMap
}

func cachedIsInputActive() *bool {
	return nil
}

func cacheInputActive() {
}
*/
