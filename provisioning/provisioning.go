package provisioning

import (
	"sync"

	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/device-monitoring/socket"
)

var (
	socketManager     *socket.Manager
	socketManagerInit sync.Once
)

// SocketManager returns the manager for the provisioning websocket
func SocketManager() *socket.Manager {
	socketManagerInit.Do(func() {
		socketManager = socket.NewManager()
		eventHandler := &eventHandler{}

		socketManager.SetEventHandler(eventHandler)
	})

	return socketManager
}

type eventHandler struct {
}

func (e *eventHandler) OnClientConnect(sendToClient chan events.Event) {
}

func (e *eventHandler) OnEventReceived(event events.Event, sendToAll chan events.Event) {
}
