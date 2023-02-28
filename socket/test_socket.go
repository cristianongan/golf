package socket

import (
	"bytes"
	"log"
	"net/http"
	"start/utils"
	"time"

	"github.com/gorilla/websocket"
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

var upgrader1 = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn
	// mu   sync.Mutex
	// Buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump() {
	defer func() {
		hubBroadcastSocket.Unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(utils.GetTimeNow().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(utils.GetTimeNow().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// log.Printf("error: %v", err)
				log.Println("[SOCKET] ReadPump err", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		hubBroadcastSocket.Broadcast <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		// select {
		// case message, ok := <-c.send:
		// 	if !ok {
		// 		log.Println("WritePump Message Error: ???")
		// 		c.write(websocket.CloseMessage, []byte{})
		// 		return
		// 	}
		// 	if err := c.write(websocket.TextMessage, message); err != nil {
		// 		log.Println("WritePump Message: ", err)
		// 		return
		// 	}
		// case <-ticker.C:
		// 		return
		// 	}
		// }
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				log.Println("[SOCKET] WritePump err The hub closed the channel.")
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Println("[SOCKET] WritePump err0 ", err)
				return
			}
			w.Write(message)

			// log.Println("[SOCKET] WritePump message ", string(message))

			// Add queued chat messages to the current websocket message.
			// n := len(c.send)
			// for i := 0; i < n; i++ {
			// 	// for msg := range c.send {
			// 	// 	msgByte, _ := json.Marshal(msg)
			// 	// 	_, err := w.Write(msgByte)
			// 	// 	if err != nil {
			// 	// 		hubBroadcastSocket.Unregister <- c
			// 	// 		break
			// 	// 	}
			// 	// }
			// 	w.Write(newline)
			// 	w.Write(<-c.send)
			// }

			if err := w.Close(); err != nil {
				// log.Println("[SOCKET] WritePump err 1", err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				// log.Println("[SOCKET] WritePump <-ticker.C err", err)
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader1.Upgrade(w, r, nil)
	if err != nil {
		log.Println("[SOCKET] ServeWs err", err)
		return
	}
	client := &Client{hub: hubBroadcastSocket, conn: conn, send: make(chan []byte, 256)}
	client.hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	// go client.ReadPump()
}

// func (c *Client) write(mt int, payload []byte) error {
// 	// c.conn.SetWriteDeadline(time.Now().Add(writeWait))
// 	c.mu.Lock()
// 	defer c.mu.Unlock()
// 	return c.conn.WriteMessage(mt, payload)
// }
