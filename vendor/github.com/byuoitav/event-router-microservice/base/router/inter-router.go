package router

import (
	"github.com/byuoitav/event-router-microservice/base"
	"github.com/byuoitav/event-router-microservice/base/node"
)

type RouterBridge struct {
	Node   *node.Node
	Router *Router
}

func StartBridge(address string, filters []string, router *Router) (*RouterBridge, error) {
	toReturn := &RouterBridge{}
	toReturn.Node = &node.Node{}
	toReturn.Router = router

	err := toReturn.Node.Start(address, filters, address)
	return toReturn, err
}

func (r *RouterBridge) ReadPassthrough() {
	for {
		msg := r.Node.Read()
		r.Router.inChan <- msg
	}
}

func (r *RouterBridge) WritePassthrough(msg base.Message) {
	r.Node.Write(msg)
}
