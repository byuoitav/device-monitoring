package eventinfrastructure

import (
	"encoding/json"

	"github.com/byuoitav/event-router-microservice/base"
	"github.com/byuoitav/event-router-microservice/base/node"
)

type EventNode struct {
	Name string

	Node *node.Node
}

// filters: an array of strings to filter events recieved by
// port: a unique port to publish events on
// addrs: addresses of subscriber to subscribe to
// name: name of event node
func NewEventNode(name string, filters []string, address string) *EventNode {
	var n EventNode
	n.Name = name
	n.Node = &node.Node{}

	n.Node.Start(address, filters, name)

	return &n
}

func (n *EventNode) PublishEvent(e Event, eventType string) error {
	return n.PublishJSONMessageByEventType(eventType, e)
}

func (n *EventNode) PublishMessageByEventType(eventType string, body []byte) {
	n.Node.Write(base.Message{MessageHeader: eventType, MessageBody: body})
}

func (n *EventNode) PublishJSONMessageByEventType(eventType string, i interface{}) error {
	toSend, err := json.Marshal(i)
	if err != nil {
		return err
	}

	n.Node.Write(base.Message{MessageHeader: eventType, MessageBody: toSend})
	return nil
}

func (n *EventNode) PublishMessage(m base.Message) {
	n.Node.Write(m)
}

func (n *EventNode) Read() base.Message {
	return n.Node.Read()
}
