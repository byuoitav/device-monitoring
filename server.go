package main

import (
	"net/http"

	"github.com/byuoitav/device-monitoring-microservice/devices"
	"github.com/labstack/echo"
)

func main() {

	devices.PingDevices()

	//	port := ":10000"
	//	router := echo.New()
	//	router.Pre(middleware.RemoveTrailingSlash())
	//	router.Use(middleware.CORS())
	//
	//	// Use the `secure` routing group to require authentication
	//	secure := router.Group("", echo.WrapMiddleware(authmiddleware.Authenticate))
	//
	//	// GET requests
	//	secure.GET("/health", Health)
	//
	//	server := http.Server{
	//		Addr:           port,
	//		MaxHeaderBytes: 1024 * 10,
	//	}
	//
	//	router.StartServer(&server)
}

func Health(context echo.Context) error {
	return context.JSON(http.StatusOK, "The fleet has moved out of lightspeed and we're preparing to - augh!")
}
