package socket

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/byuoitav/device-monitoring/model"
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
	sendChan chan model.Event
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
	var event model.Event
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			slog.Warn("error reading from websocket",
				slog.String("address",
					c.conn.RemoteAddr().String()),
				slog.String("error", err.Error()))
			break
		}

		err = json.Unmarshal(message, &event)
		if err != nil {
			slog.Warn("unable to unmarshal event from client",
				slog.String("address", c.conn.RemoteAddr().String()),
				slog.String("error", err.Error()))
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
			for range n {
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
