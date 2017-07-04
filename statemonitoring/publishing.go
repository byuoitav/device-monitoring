package statemonitoring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
	"github.com/byuoitav/event-router-microservice/subscription"
	"github.com/xuther/go-message-router/common"
	"github.com/xuther/go-message-router/publisher"
)

var Publisher publisher.Publisher

func StartPublisher() {
	//Start our publisher for publishing satate events
	var err error
	Publisher, err = publisher.NewPublisher("7004", 1000, 10)
	if err != nil {
		errstr := fmt.Sprintf("[publisher] Could not start publisher. Error: %v\n", err.Error())
		log.Fatalf(errstr)
	}

	go func() {
		Publisher.Listen()
		if err != nil {
			errstr := fmt.Sprintf("[publisher] Could not start publisher listening. Error: %v\n", err.Error())
			log.Fatalf(errstr)
		} else {
			log.Printf("[publisher] Publisher started on port :7004")
		}
	}()

	if len(os.Getenv("LOCAL_ENVIRONMENT")) > 0 {
		go func() {
			var s subscription.SubscribeRequest
			s.Address = "localhost:7004"
			body, err := json.Marshal(s)
			if err != nil {
				log.Printf("[error] %s", err.Error())
			}
			_, err = http.Post("http://localhost:6999/subscribe", "application/json", bytes.NewBuffer(body))

			for err != nil {
				_, err = http.Post("http://localhost:6999/subscribe", "application/json", bytes.NewBuffer(body))
				log.Printf("[error] The router hasn't subscribed to me yet. Trying again...")
				time.Sleep(3 * time.Second)
			}
			log.Printf("[publisher] Router is subscribed to me")
		}()
	}
}

func SendEvent(Type eventinfrastructure.EventType,
	Cause eventinfrastructure.EventCause,
	Device string,
	Room string,
	Building string,
	InfoKey string,
	InfoValue string,
	Error bool) error {

	e := eventinfrastructure.EventInfo{
		Type:           Type,
		EventCause:     Cause,
		Device:         Device,
		EventInfoKey:   InfoKey,
		EventInfoValue: InfoValue,
	}

	err := Publish(eventinfrastructure.Event{
		Event:    e,
		Building: Building,
		Room:     Room,
	}, Error)

	return err
}

func PublishError(errorStr string, cause eventinfrastructure.EventCause) {
	e := eventinfrastructure.EventInfo{
		Type:           eventinfrastructure.ERROR,
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
		}
	}

	Publish(eventinfrastructure.Event{
		Event:    e,
		Building: building,
		Room:     room,
	}, true)
}

func Publish(e eventinfrastructure.Event, Error bool) error {
	var err error

	// create the event
	e.Timestamp = time.Now().Format(time.RFC3339)
	if len(os.Getenv("LOCAL_ENVIRONMENT")) > 0 {
		e.Hostname = os.Getenv("PI_HOSTNAME")
		if len(os.Getenv("DEVELOPMENT_HOSTNAME")) > 0 {
			e.Hostname = os.Getenv("DEVELOPMENT_HOSTNAME")
		}
	} else {
		// isn't it running in a docker container in aws? this won't work?
		e.Hostname, err = os.Hostname()
	}
	if err != nil {
		return err
	}

	e.LocalEnvironment = len(os.Getenv("LOCAL_ENVIRONMENT")) > 0

	toSend, err := json.Marshal(&e)
	if err != nil {
		return err
	}

	header := [24]byte{}
	if !Error {
		copy(header[:], eventinfrastructure.APISuccess)
	} else {
		copy(header[:], eventinfrastructure.APIError)
	}

	log.Printf("Publishing event: %s", toSend)
	Publisher.Write(common.Message{MessageHeader: header, MessageBody: toSend})

	return err
}
