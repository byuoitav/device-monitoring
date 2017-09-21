package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/byuoitav/authmiddleware"
	"github.com/byuoitav/device-monitoring-microservice/device"
	"github.com/byuoitav/device-monitoring-microservice/handlers"
	"github.com/byuoitav/device-monitoring-microservice/monitoring"
	"github.com/byuoitav/device-monitoring-microservice/statusinfrastructure"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
	"github.com/byuoitav/touchpanel-ui-microservice/socket"
	"github.com/fatih/color"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var addr string
var building string
var room string

func main() {
	// start event node
	filters := []string{eventinfrastructure.TestEnd, eventinfrastructure.TestReply, eventinfrastructure.TestExternalReply}
	en := eventinfrastructure.NewEventNode("Device Monitoring", "7004", filters, os.Getenv("EVENT_ROUTER_ADDRESS"))

	// websocket
	hub := socket.NewHub()
	go WriteEventsToSocket(en, hub, statusinfrastructure.EventNodeStatus{})

	//get building and room info
	hostname := os.Getenv("PI_HOSTNAME")
	building = strings.Split(hostname, "-")[0]
	room = strings.Split(hostname, "-")[1]

	go monitor(building, room, en)

	//get addresses from database
	devices, err := device.GetAddresses(building, room)
	if err != nil {
		log.Printf("Error getting devices from database: %s", err.Error())
	}

	//figure out how often to ping devices and start process in new goroutine
	pingInterval := os.Getenv("DEVICE_PING_INTERVAL")
	interval, err := strconv.Atoi(pingInterval)
	if err != nil {
		log.Printf("Error reading check interval. Terminating...")
	} else {
		go func() {
			for {
				device.PingAddresses(building, room, devices)
				time.Sleep(time.Duration(interval) * time.Second)
			}
		}()
	}

	port := ":10000"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	secure := router.Group("", echo.WrapMiddleware(authmiddleware.Authenticate))

	// websocket
	router.GET("/websocket", func(context echo.Context) error {
		socket.ServeWebsocket(hub, context.Response().Writer, context.Request())
		return nil
	})

	secure.GET("/health", handlers.Health)
	secure.GET("/pulse", Pulse)
	secure.GET("/eventstatus", handlers.EventStatus, BindEventNode(en))
	secure.GET("/testevents", func(context echo.Context) error {
		en.PublishMessageByEventType(eventinfrastructure.TestStart, []byte("test event"))
		return nil
	})

	secure.GET("/hostname", handlers.GetHostname)
	secure.GET("/ip", handlers.GetIP)
	secure.GET("/network", handlers.GetNetworkConnectedStatus)

	secure.GET("/reboot", handlers.RebootPi)

	router.Static("/dash", "dash")

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	router.StartServer(&server)

}

func Pulse(context echo.Context) error {
	err := monitoring.GetAndReportStatus(addr, building, room)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, err.Error())
	}

	return context.JSON(http.StatusOK, "Pulse sent.")
}

func BindEventNode(en *eventinfrastructure.EventNode) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(eventinfrastructure.ContextEventNode, en)
			return next(c)
		}
	}
}

func monitor(building, room string, en *eventinfrastructure.EventNode) {
	currentlyMonitoring := false

	for {
		shouldIMonitor := monitoring.ShouldIMonitorAPI()

		if shouldIMonitor && !currentlyMonitoring {
			color.Set(color.FgYellow, color.Bold)
			log.Printf("Starting monitoring of API")
			color.Unset()
			addr = monitoring.StartMonitoring(time.Second*300, "localhost:8000", building, room, en)
			currentlyMonitoring = true
		} else if currentlyMonitoring && shouldIMonitor {
		} else {
			color.Set(color.FgYellow, color.Bold)
			log.Printf("Stopping monitoring of API")
			color.Unset()

			// stop monitoring?
			monitoring.StopMonitoring()
			currentlyMonitoring = false
		}
		time.Sleep(time.Second * 15)
	}
}

func WriteEventsToSocket(en *eventinfrastructure.EventNode, h *socket.Hub, t interface{}) {
	for {
		select {
		case message, ok := <-en.Read:
			if !ok {
				color.Set(color.FgRed)
				log.Fatalf("eventnode read channel closed.")
				color.Unset()
			}

			header := string(bytes.Trim(message.MessageHeader[:], "\x00"))
			if strings.EqualFold(header, eventinfrastructure.TestExternal) {
				color.Set(color.FgBlue, color.Bold)
				log.Printf("Responding to external test event")
				color.Unset()

				var s statusinfrastructure.EventNodeStatus
				if len(os.Getenv("DEVELOPMENT_HOSTNAME")) > 0 {
					s.Name = os.Getenv("DEVELOPMENT_HOSTNAME")
				} else if len(os.Getenv("PI_HOSTNAME")) > 0 {
					s.Name = os.Getenv("PI_HOSTNAME")
				} else {
					s.Name, _ = os.Hostname()
				}

				en.PublishJSONMessageByEventType(eventinfrastructure.TestExternalReply, s)
			}

			err := json.Unmarshal(message.MessageBody, &t)
			if err != nil {
				color.Set(color.FgRed)
				log.Printf("failed to unmarshal message into Event type: %s", message.MessageBody)
				color.Unset()
			} else {
				h.WriteToSockets(t)
			}
		}
	}
}
