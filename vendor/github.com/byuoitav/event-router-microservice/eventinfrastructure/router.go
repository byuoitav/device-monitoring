package eventinfrastructure

import (
	"log"

	"github.com/byuoitav/event-router-microservice/base/router"
	"github.com/fatih/color"
)

func NewRouter(routingTable map[string][]string, addrs []string) (*router.Router, error) {

	r := router.NewRouter()

	go r.StartRouter(routingTable)

	err := r.ConnectToRouters(addrs, routingTable)
	if err != nil {
		log.Printf(color.HiRedString("Could not connect to peers: %v", err.Error()))
		return r, err
	}

	return r, nil
}
