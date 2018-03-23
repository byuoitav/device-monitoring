package main

import (
	"log"
	"net/http"

	"github.com/byuoitav/event-router-microservice/base/router"
	ei "github.com/byuoitav/event-router-microservice/eventinfrastructure"
	"github.com/byuoitav/event-router-microservice/helpers"
	"github.com/fatih/color"
	"github.com/jessemillar/health"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	defer color.Unset()

	RoutingTable := buildTable()

	//pretty print the routing table
	helpers.PrettyPrint(RoutingTable)

	// create the router
	addrs := helpers.GetOutsideAddresses()
	route, err := ei.NewRouter(RoutingTable, addrs)
	if err != nil {
		log.Printf(color.HiRedString("Could not create router. Error: %v", err.Error()))
		return
	}
	log.Printf(color.HiGreenString("Router Started... Starting server."))

	server := echo.New()
	server.Pre(middleware.RemoveTrailingSlash())
	server.Use(middleware.CORS())

	server.GET("/health", echo.WrapHandler(http.HandlerFunc(health.Check)))
	server.GET("/mstatus", func(context echo.Context) error {
		return helpers.GetStatus(context, route)
	})
	server.GET("/messagelogs/:val", func(context echo.Context) error {
		return helpers.SetMessageLogLevel(route, context)
	})

	server.GET("/subscribe", func(context echo.Context) error {
		return router.ListenForNodes(route, context)
	})

	server.Start(":7000")
}

func buildTable() map[string][]string {

	RoutingTable := make(map[string][]string)
	RoutingTable[ei.TestStart] = []string{ei.TestPleaseReply}    // local DM  --> local microservices (everyone listens to TestPleaseReply) and external routers
	RoutingTable[ei.TestPleaseReply] = []string{ei.TestExternal} // external routers --> external DM's
	RoutingTable[ei.TestExternalReply] = []string{ei.TestReply}  // external DM --> external router
	RoutingTable[ei.TestReply] = []string{ei.TestEnd}            // local microservices and external DM --> local DM

	RoutingTable[ei.Room] = []string{ei.UI}
	RoutingTable[ei.APISuccess] = []string{
		ei.Translator,
		ei.UI,
		ei.Room,
	}
	RoutingTable[ei.External] = []string{ei.UI}
	RoutingTable[ei.APIError] = []string{ei.UI, ei.Translator}
	RoutingTable[ei.Metrics] = []string{ei.Translator}
	RoutingTable[ei.UIFeature] = []string{ei.Room}
	RoutingTable[ei.RoomDivide] = []string{
		ei.Translator,
		ei.UI,
		ei.Room,
	}

	return RoutingTable
}
