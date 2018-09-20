package monitoring

/*

func SendEvent(Type events.EventType,
	Cause events.EventCause,
	Device string,
	DeviceID string,
	Room string,
	Building string,
	InfoKey string,
	InfoValue string,
	Error bool) error {

	e := events.EventInfo{
		Type:           Type,
		EventCause:     Cause,
		Device:         Device,
		DeviceID:       DeviceID,
		EventInfoKey:   InfoKey,
		EventInfoValue: InfoValue,
	}

	err := Publish(events.Event{
		Event:    e,
		Building: Building,
		Room:     Room,
	}, Error)

	return err
}

func PublishError(errorStr string, cause events.EventCause) {
	e := events.EventInfo{
		Type:           events.ERROR,
		EventCause:     cause,
		EventInfoKey:   "Error String",
		EventInfoValue: errorStr,
	}

	building := ""
	room := ""

	if len(os.Getenv("LOCAL_ENVIRONMENT")) > 0 {
		if len(os.Getenv("PI_HOSTNAME")) > 0 {
			name := os.Getenv("PI_HOSTNAME")
			roomInfo := strings.Split(name, "-")
			building = roomInfo[0]
			room = roomInfo[1]
			e.Device = roomInfo[2]
			e.DeviceID = name
		}
	}

	Publish(events.Event{
		Event:    e,
		Building: building,
		Room:     room,
	}, true)
}

func Publish(e events.Event, Error bool) error {
	var err error

	// create the event
	e.Timestamp = time.Now().Format(time.RFC3339)
	if len(os.Getenv("LOCAL_ENVIRONMENT")) > 0 {
		e.Hostname = os.Getenv("PI_HOSTNAME")
		if len(os.Getenv("DEVELOPMENT_HOSTNAME")) > 0 {
			e.Hostname = os.Getenv("DEVELOPMENT_HOSTNAME")
		}
	} else {
		e.Hostname, err = os.Hostname()
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}

	e.LocalEnvironment = len(os.Getenv("LOCAL_ENVIRONMENT")) > 0

	if !Error {
		eventnode.PublishEvent(events.APISuccess, e)
	} else {
		eventnode.PublishEvent(events.APIError, e)
	}

	return err
}
*/
