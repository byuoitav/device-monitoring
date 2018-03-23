package base

import (
	"time"

	"github.com/xuther/go-message-router/common"
)

type Router interface {
	StartRouter(RoutingTable map[string][]string) error
	ConnectToRouters([]string) error
	Stop() error
}

type Node interface {
	GetState() //will return something
	ConnectToRouter(address string, filters []string) error
	Write(common.Message) error
	Read() common.Message
	Close()
}

type subscriptionReq struct {
	Address  string
	Count    int
	Interval time.Duration
}
type Message struct {
	MessageHeader string `json:"message-header"` //Header is the event type
	MessageBody   []byte `json:"message-body"`   //Body can be whatever message is desired.
}
