package realtime

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/exam-arena/internal/platform/tokens"
	"github.com/gorilla/websocket"
)

// ConnectHandler turns an HTTP request into a live WebSocket connection
// for a logged-in player, and answers the built-in "ping" message.
// Game messages (answers, accepting a match…) are registered on the hub by
// the features that own them, e.g. match.RegisterGameMessages.
type ConnectHandler struct {
	hub       *Hub
	jwtSecret string
}

func NewConnectHandler(hub *Hub, jwtSecret string) *ConnectHandler {
	h := &ConnectHandler{hub: hub, jwtSecret: jwtSecret}
	hub.RegisterHandler("ping", h.handlePing)
	return h
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // restrict to your domain in production
	},
}

// HandleConnection upgrades the HTTP connection to WebSocket.
//
// Auth: the JWT is passed as a query parameter because browsers
// cannot set custom headers on the WebSocket handshake request.
//
//	ws://host/ws?token=<JWT>
func (h *ConnectHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token query parameter", http.StatusUnauthorized)
		return
	}

	claims, err := tokens.Validate(token, h.jwtSecret)
	if err != nil {
		http.Error(w, "invalid or expired token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}

	client := NewClient(h.hub, conn, claims.UserID, claims.Username)
	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()

	// Confirm the connection to the client.
	client.SendJSON(Message{
		Type: "connected",
		Payload: map[string]interface{}{
			"user_id":  claims.UserID,
			"username": claims.Username,
			"message":  "connected to Exam Arena",
		},
	})

	slog.Info("ws client connected",
		"user_id", claims.UserID,
		"username", claims.Username,
	)
}

// handlePing responds with a pong so the client can measure latency.
//
// Client sends:  { "type": "ping", "payload": {} }
// Server responds: { "type": "pong", "payload": { "server_time": "..." } }
func (h *ConnectHandler) handlePing(client *Client, _ json.RawMessage) {
	client.SendJSON(Message{
		Type: "pong",
		Payload: map[string]interface{}{
			"server_time": "ok",
		},
	})
}
