package socket

import (
	"fmt"
	"net/http"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/v2/events"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
)

type (
	// An EventHandler is a struct that handles websocket events
	EventHandler interface {
		// OnClientConnect is called once each time a new client is connected.
		// use sendToClient to send events to the new client.
		OnClientConnect(sendToClient chan events.Event)

		// OnEventReceived will be called each time _any_ client sends an event.
		// event is the event recieved, and events can be sent back to _all_ clients using sendToAll.
		OnEventReceived(event events.Event, sendToAll chan events.Event)
	}

	// A Manager manages a group of websocket connections
	Manager struct {
		clients    map[*Client]bool
		register   chan *Client
		unregister chan *Client

		broadcast    chan events.Event
		eventHandler EventHandler
	}
)

// NewManager returns new manager.
func NewManager() *Manager {
	m := &Manager{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),

		broadcast: make(chan events.Event),
	}

	go m.run()
	return m
}

// SetEventHandler sets the event handler for the websocket manager
func (m *Manager) SetEventHandler(handler EventHandler) {
	m.eventHandler = handler
}

func (m *Manager) run() {
	for {
		select {
		case client := <-m.register:
			log.L.Infof("Registering %s to websocket manager", client.conn.RemoteAddr())
			m.clients[client] = true

			if m.eventHandler != nil {
				m.eventHandler.OnClientConnect(client.sendChan)
			}
		case client := <-m.unregister:
			if _, ok := m.clients[client]; ok {
				log.L.Infof("Removing %s from websocket manager", client.conn.RemoteAddr())
				close(client.sendChan)
				delete(m.clients, client)
			}
		case message := <-m.broadcast:
			log.L.Debugf("broadcasting message to %v clients: %s", len(m.clients), message)
			for client := range m.clients {
				select {
				case client.sendChan <- message:
				default:
					m.unregister <- client
				}
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// UpgradeToWebsocket upgrades a connection to a websocket and creates a client for the connection
func UpgradeToWebsocket(manager *Manager) func(ctx echo.Context) error {
	return func(ctx echo.Context) error {
		log.L.Infof("Attempting to uppgrading HTTP connection from %s to websocket", ctx.Request().RemoteAddr)
		conn, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
		if err != nil {
			return ctx.String(http.StatusInternalServerError, fmt.Sprintf("unable to upgrade connection to a websocket: %s", err))
		}

		client := &Client{
			manager:  manager,
			conn:     conn,
			sendChan: make(chan events.Event, 256),
		}
		client.manager.register <- client

		go client.writePump()
		go client.readPump()

		return nil
	}
}
