package common

type Message struct {
	MessageHeader [24]byte `json:"message-header"` //Header is the event type
	MessageBody   []byte   `json:"message-body"`   //Body can be whatever message is desired.
}
