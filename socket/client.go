package socket

import (
	"encoding/json"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/v2/events"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// A Client manages a specific websocket connection to a client
type Client struct {
	manager  *Manager
	conn     *websocket.Conn
	sendChan chan events.Event
}

func (c *Client) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// read messages from client forever
	var event events.Event
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.L.Warnf("error reading from %s's websocket: %v", c.conn.RemoteAddr(), err)
			break
		}

		err = json.Unmarshal(message, &event)
		if err != nil {
			log.L.Warnf("unable to unmarshal event from client %s: %v", c.conn.RemoteAddr(), err)
			break
		}

		if c.manager.eventHandler != nil {
			c.manager.eventHandler.OnEventReceived(event, c.manager.broadcast)
		}
		// message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case event, ok := <-c.sendChan:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			// marshal event
			bytes, err := json.Marshal(event)
			if err != nil {
				return
			}
			w.Write(bytes)

			n := len(c.sendChan)
			for i := 0; i < n; i++ {
				w.Write(newline)

				bytes, err := json.Marshal(<-c.sendChan)
				if err != nil {
					return
				}
				w.Write(bytes)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
