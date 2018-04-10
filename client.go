//Simple Websock Library for Go
package sock

import (
	"bytes"

	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// MessageBlock is a struct containing a message and a *Client of the client that sent the message
type MessageBlock struct {
	//Client that sent Message
	Client *Client
	//The message sent to the server
	Message string
}

// Client is a connection between the server and client
type Client struct {
	//Hub where Client is located
	hub *Hub
	//Use to distinguish connections.
	//Default is 1
	Id int
	// The websocket Connection.
	Conn *websocket.Conn

	// Channel to send to client
	Send chan []byte
}

// readPump pumps messages from the websocket Connection to the hub.
//
// The application runs readPump in a per-Connection goroutine. The application
// ensures that there is at most one reader on a Connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {

		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		//fmt.Print(message)
		c.hub.Messages[len(c.hub.Messages)] = string(message)
		//fmt.Println(c.hub.messages[len(c.hub.messages)-1])
		input := MessageBlock{Client: c, Message: c.hub.Messages[len(c.hub.Messages)-1]}

		c.hub.Input <- input
	}
}

// writePump pumps messages from the hub to the websocket Connection.
//
// A goroutine running writePump is started for each Connection. The
// application ensures that there is at most one writer to a Connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs initializes connection between hub and client
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	Conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, Conn: Conn, Send: make(chan []byte, 256), Id: 1}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
