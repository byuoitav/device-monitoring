package main

import (
	"net/http"

	"github.com/byuoitav/authmiddleware"
	"github.com/byuoitav/common"
	"github.com/byuoitav/device-monitoring-microservice/handlers"
	"github.com/byuoitav/device-monitoring-microservice/jobs"
	"github.com/labstack/echo"
)

func main() {
	// start jobs
	go jobs.StartJobScheduler()

	// server
	port := ":10000"
	router := common.NewRouter()

	secure := router.Group("", echo.WrapMiddleware(authmiddleware.Authenticate))

	// device info endpoints
	secure.GET("/device", handlers.GetDeviceInfo)
	secure.GET("/device/hostname", handlers.GetHostname)
	secure.GET("/device/ip", handlers.GetIPAddress)
	secure.GET("/device/pihostname", handlers.GetDeviceID)
	secure.GET("/device/network", handlers.IsConnectedToInternet)

	// action endpoints
	secure.PUT("/reboot", handlers.RebootPi)

	// dashboard
	secure.Static("/dash", "dash-dist")

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}
	router.StartServer(&server)

	/*
		// websocket
		hub := socket.NewHub(en)
		go WriteEventsToSocket(en, hub, statusinfrastructure.EventNodeStatus{})

		port := ":10000"
		// websocket
		router.GET("/websocket", func(context echo.Context) error {
			socket.ServeWebsocket(hub, context.Response().Writer, context.Request())
			return nil
		})

		secure.GET("/pulse", Pulse)
		secure.GET("/eventstatus", handlers.EventStatus, BindEventNode(en))
		secure.GET("/testevents", func(context echo.Context) error {
			en.Node.Write(messenger.Message{Header: events.TestStart, Body: []byte("test event")})
			return nil
		})
	*/
}

/*
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
*/
