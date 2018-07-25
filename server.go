package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/byuoitav/authmiddleware"
	"github.com/byuoitav/common/events"
	"github.com/byuoitav/device-monitoring-microservice/device"
	"github.com/byuoitav/device-monitoring-microservice/handlers"
	"github.com/byuoitav/device-monitoring-microservice/monitoring"
	"github.com/byuoitav/device-monitoring-microservice/statusinfrastructure"
	"github.com/byuoitav/messenger"
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
	filters := []string{events.TestEnd, events.TestExternal}
	en := events.NewEventNode("Device Monitoring", os.Getenv("EVENT_ROUTER_ADDRESS"), filters)

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
	router.Use(echo.WrapMiddleware(authmiddleware.Authenticate))

	secure := router.Group("", echo.WrapMiddleware(authmiddleware.AuthenticateUser))

	// websocket
	router.GET("/websocket", func(context echo.Context) error {
		socket.ServeWebsocket(hub, context.Response().Writer, context.Request())
		return nil
	})

	secure.GET("/health", handlers.Health)
	secure.GET("/pulse", Pulse)
	secure.GET("/eventstatus", handlers.EventStatus, BindEventNode(en))
	secure.GET("/testevents", func(context echo.Context) error {
		en.Node.Write(messenger.Message{Header: events.TestStart, Body: []byte("test event")})
		return nil
	})

	router.GET("/hostname", handlers.GetHostname)
	router.GET("/ip", handlers.GetIP)
	router.GET("/network", handlers.GetNetworkConnectedStatus)

	secure.GET("/reboot", handlers.RebootPi)

	secure.Static("/dash", "dash-dist")

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

func BindEventNode(en *events.EventNode) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(events.ContextEventNode, en)
			return next(c)
		}
	}
}

func monitor(building, room string, en *events.EventNode) {
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

func WriteEventsToSocket(en *events.EventNode, h *socket.Hub, t interface{}) {
	for {
		message := en.Node.Read()

		if strings.EqualFold(message.Header, events.TestExternal) {
			log.Printf(color.BlueString("Responding to external test event"))

			var s statusinfrastructure.EventNodeStatus
			if len(os.Getenv("DEVELOPMENT_HOSTNAME")) > 0 {
				s.Name = os.Getenv("DEVELOPMENT_HOSTNAME")
			} else if len(os.Getenv("PI_HOSTNAME")) > 0 {
				s.Name = os.Getenv("PI_HOSTNAME")
			} else {
				s.Name, _ = os.Hostname()
			}

			b, err := json.Marshal(s)
			if err != nil {
				log.Printf("error marshaling json: %v", err.Error())
				continue
			}

			en.Node.Write(messenger.Message{Header: events.TestExternalReply, Body: b})
		}

		err := json.Unmarshal(message.Body, &t)
		if err != nil {
			log.Printf(color.RedString("failed to unmarshal message into Event type: %s", message.Body))
		} else {
			h.WriteToSockets(t)
		}
	}
}
