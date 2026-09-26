package ws

import (
	"encoding/json"
	"log/slog"
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
				close(existing.send)
				delete(h.clients, client.userID)
			}
			h.clients[client.userID] = client
			h.mu.Unlock()
			slog.Info("client connected", "user_id", client.userID, "total", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if existing, ok := h.clients[client.userID]; ok && existing == client {
				delete(h.clients, client.userID)
				close(client.send)
			}
			h.mu.Unlock()
			slog.Info("client disconnected", "user_id", client.userID, "total", len(h.clients))

		case msg := <-h.incoming:
			h.handleMessage(msg)

		case <-h.done:
			h.mu.Lock()
			for id, client := range h.clients {
				close(client.send)
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

func (h *Hub) RegisterHandler(msgType string, handler MessageHandler) {
	h.handlers[msgType] = handler
}

func (h *Hub) handleMessage(msg *IncomingMessage) {
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
		select {
		case client.send <- data:
		default:
		}
	}
}

func (h *Hub) IsOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

func (h *Hub) OnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) Shutdown() {
	close(h.done)
}
