package realtime

import (
	"encoding/json"
	"log/slog"
	"runtime/debug"
	"sync"
)

type Hub struct {
	// Registered clients indexed by user ID for O(1) lookup
	clients    map[string]*Client
	mu         sync.RWMutex
	register   chan *Client
	unregister chan *Client
	incoming   chan *IncomingMessage
	handlers   map[string]MessageHandler
	done       chan struct{}
	// onPresence, if set, is called (in its own goroutine) when a user
	// connects or disconnects.
	onPresence func(userID string, online bool)
}

type MessageHandler func(client *Client, payload json.RawMessage)

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
		incoming:   make(chan *IncomingMessage, 256),
		handlers:   make(map[string]MessageHandler),
		done:       make(chan struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// Close existing connection for this user (single session)
			if existing, ok := h.clients[client.userID]; ok {
				existing.closeSend()
				delete(h.clients, client.userID)
			}
			h.clients[client.userID] = client
			total := len(h.clients)
			h.mu.Unlock()
			slog.Info("client connected", "user_id", client.userID, "total", total)
			h.notifyPresence(client.userID, true)

		case client := <-h.unregister:
			h.mu.Lock()
			removed := false
			if existing, ok := h.clients[client.userID]; ok && existing == client {
				delete(h.clients, client.userID)
				client.closeSend()
				removed = true
			}
			total := len(h.clients)
			h.mu.Unlock()
			slog.Info("client disconnected", "user_id", client.userID, "total", total)
			if removed {
				h.notifyPresence(client.userID, false)
			}

		case msg := <-h.incoming:
			h.handleMessage(msg)

		case <-h.done:
			h.mu.Lock()
			for id, client := range h.clients {
				client.closeSend()
				delete(h.clients, id)
			}
			h.mu.Unlock()
			return
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

// SetPresenceHook must be called before Run.
func (h *Hub) SetPresenceHook(fn func(userID string, online bool)) {
	h.onPresence = fn
}

func (h *Hub) notifyPresence(userID string, online bool) {
	if h.onPresence != nil {
		go h.onPresence(userID, online)
	}
}

// RegisterHandler must be called before Run starts consuming messages.
func (h *Hub) RegisterHandler(msgType string, handler MessageHandler) {
	h.handlers[msgType] = handler
}

// handleMessage dispatches one inbound frame. It runs on the hub goroutine,
// which has no HTTP recovery middleware above it, so a panic inside a handler
// is recovered here — one malformed message must not take the server down.
func (h *Hub) handleMessage(msg *IncomingMessage) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic in websocket handler",
				"error", r,
				"user_id", msg.Client.userID,
				"stack", string(debug.Stack()),
			)
			msg.Client.SendJSON(Message{
				Type:    "error",
				Payload: map[string]interface{}{"message": "internal server error"},
			})
		}
	}()

	var parsed struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}

	if err := json.Unmarshal(msg.Payload, &parsed); err != nil {
		slog.Error("failed to parse ws message", "error", err)
		return
	}

	if handler, ok := h.handlers[parsed.Type]; ok {
		handler(msg.Client, parsed.Payload)
	} else {
		slog.Warn("unknown message type", "type", parsed.Type)
	}
}

func (h *Hub) SendToUser(userID string, msg Message) {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()

	if ok {
		client.SendJSON(msg)
	}
}

func (h *Hub) Broadcast(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		client.enqueue(data)
	}
}

func (h *Hub) IsOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

// OnlineUserIDs returns the IDs of every connected user.
func (h *Hub) OnlineUserIDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ids := make([]string, 0, len(h.clients))
	for id := range h.clients {
		ids = append(ids, id)
	}
	return ids
}

func (h *Hub) OnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) Shutdown() {
	close(h.done)
}
