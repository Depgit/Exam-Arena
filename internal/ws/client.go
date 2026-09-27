package ws

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   string
	username string

	// mu guards closed and every send on the channel. The hub closes the
	// channel when a user reconnects (single session per user) while other
	// goroutines — the match timer, endMatch, the matchmaking engine — may
	// still hold a reference to this client and try to write to it. Without
	// the guard that write panics with "send on closed channel" and, since
	// those goroutines have no recover, takes the whole process down.
	mu     sync.Mutex
	closed bool
}

func NewClient(hub *Hub, conn *websocket.Conn, userID, username string) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID,
		username: username,
	}
}

func (c *Client) UserID() string {
	return c.userID
}

func (c *Client) Username() string {
	return c.username
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("websocket read error", "error", err, "user", c.userID)
			}
			break
		}

		// Forward message to hub for processing
		c.hub.incoming <- &IncomingMessage{
			Client:  c,
			Payload: message,
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Drain queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.send)
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

func (c *Client) SendJSON(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal message", "error", err)
		return
	}
	c.enqueue(data)
}

// enqueue queues an already-encoded frame without blocking. It returns false
// when the client has been closed or its buffer is full, in which case the
// frame is dropped.
func (c *Client) enqueue(data []byte) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return false
	}
	select {
	case c.send <- data:
		return true
	default:
		slog.Warn("client send buffer full, dropping message", "user", c.userID)
		return false
	}
}

// closeSend closes the outbound channel exactly once. WritePump sees the
// closed channel, sends a close frame and shuts the connection down.
func (c *Client) closeSend() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}
	c.closed = true
	close(c.send)
}

type IncomingMessage struct {
	Client  *Client
	Payload []byte
}
