package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/exam-arena/internal/service"
	"github.com/exam-arena/internal/utils"
	"github.com/exam-arena/internal/ws"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // restrict to your domain in production
	},
}

type WSHandler struct {
	hub          *ws.Hub
	jwtSecret    string
	matchService *service.MatchService
}

// NewWSHandler takes three arguments: the hub, JWT secret, and the
// MatchService so that WebSocket answer submissions can be routed
// directly into the live match engine.
func NewWSHandler(hub *ws.Hub, jwtSecret string, matchService *service.MatchService) *WSHandler {
	h := &WSHandler{
		hub:          hub,
		jwtSecret:    jwtSecret,
		matchService: matchService,
	}
	h.registerHandlers()
	return h
}

// registerHandlers maps every incoming WebSocket message type to the
// function that processes it. Add new types here as the game grows.
func (h *WSHandler) registerHandlers() {
	h.hub.RegisterHandler("submit_answer", h.handleSubmitAnswer)
	h.hub.RegisterHandler("accept_match", h.handleAcceptMatch)
	h.hub.RegisterHandler("decline_match", h.handleDeclineMatch)
	h.hub.RegisterHandler("ping", h.handlePing)
}

// HandleConnection upgrades the HTTP connection to WebSocket.
//
// Auth: the JWT is passed as a query parameter because browsers
// cannot set custom headers on the WebSocket handshake request.
//
//	ws://host/ws?token=<JWT>
func (h *WSHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token query parameter", http.StatusUnauthorized)
		return
	}

	claims, err := utils.ValidateToken(token, h.jwtSecret)
	if err != nil {
		http.Error(w, "invalid or expired token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}

	client := ws.NewClient(h.hub, conn, claims.UserID, claims.Username)
	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()

	// Confirm the connection to the client.
	client.SendJSON(ws.Message{
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

// ── Message Handlers ─────────────────────────────────────────────────────

// handleSubmitAnswer processes an answer submitted during a live match.
//
// Expected payload:
//
//	{
//	  "match_id":     "uuid",
//	  "question_id":  "uuid",
//	  "option_id":    "uuid",
//	  "time_taken_ms": 4321
//	}
//
// The MatchService checks correctness from the in-memory LiveMatch
// (zero DB calls), updates scores, broadcasts to all players, and
// persists the answer to Postgres asynchronously.
func (h *WSHandler) handleSubmitAnswer(client *ws.Client, payload json.RawMessage) {
	var data struct {
		MatchID     string `json:"match_id"`
		QuestionID  string `json:"question_id"`
		OptionID    string `json:"option_id"`
		TimeTakenMs int    `json:"time_taken_ms"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		client.SendJSON(ws.Message{
			Type:    "error",
			Payload: map[string]interface{}{"message": "invalid answer payload"},
		})
		return
	}

	// Validate required fields.
	if data.MatchID == "" || data.QuestionID == "" {
		client.SendJSON(ws.Message{
			Type:    "error",
			Payload: map[string]interface{}{"message": "match_id and question_id are required"},
		})
		return
	}

	// Clamp time_taken_ms to sane bounds.
	if data.TimeTakenMs < 0 {
		data.TimeTakenMs = 0
	}
	if data.TimeTakenMs > 300_000 { // 5 minutes max
		data.TimeTakenMs = 300_000
	}

	req := service.AnswerRequest{
		MatchID:     data.MatchID,
		UserID:      client.UserID(),
		QuestionID:  data.QuestionID,
		OptionID:    data.OptionID,
		TimeTakenMs: data.TimeTakenMs,
	}

	if err := h.matchService.SubmitAnswer(context.Background(), req); err != nil {
		slog.Warn("answer submission failed",
			"user_id", client.UserID(),
			"match_id", data.MatchID,
			"error", err,
		)
		client.SendJSON(ws.Message{
			Type: "error",
			Payload: map[string]interface{}{
				"message":  err.Error(),
				"match_id": data.MatchID,
			},
		})
		return
	}

	// No explicit ACK — the MatchService broadcasts a "score_update"
	// message to ALL players in the match, including the sender.
}

// handleAcceptMatch accepts a proposed match during the 30-second
// acceptance window (game_rules.md §4).
//
// Expected payload:
//
//	{ "pending_id": "..." }
//
// The MatchService re-checks that the sender is actually one of the two
// players before recording anything, so a guessed or copied pending_id gets
// a player nowhere. The match starts once both sides have accepted.
func (h *WSHandler) handleAcceptMatch(client *ws.Client, payload json.RawMessage) {
	pendingID, ok := decodePendingID(client, payload)
	if !ok {
		return
	}

	if err := h.matchService.AcceptMatch(client.UserID(), pendingID); err != nil {
		slog.Warn("match accept failed",
			"user_id", client.UserID(), "pending_id", pendingID, "error", err)
		client.SendJSON(ws.Message{
			Type:    "error",
			Payload: map[string]interface{}{"message": err.Error()},
		})
	}
}

// handleDeclineMatch turns down a proposed match. The opponent is put back
// into the matchmaking queue; the decliner is not.
func (h *WSHandler) handleDeclineMatch(client *ws.Client, payload json.RawMessage) {
	pendingID, ok := decodePendingID(client, payload)
	if !ok {
		return
	}

	if err := h.matchService.DeclineMatch(client.UserID(), pendingID); err != nil {
		// A decline arriving just after the window closed is routine, not an
		// error worth putting in front of the player.
		slog.Info("match decline ignored",
			"user_id", client.UserID(), "pending_id", pendingID, "reason", err)
	}
}

// decodePendingID pulls the pending match id out of an accept/decline
// payload, replying with an error frame if it is missing or malformed.
func decodePendingID(client *ws.Client, payload json.RawMessage) (string, bool) {
	var data struct {
		PendingID string `json:"pending_id"`
	}
	if err := json.Unmarshal(payload, &data); err != nil || data.PendingID == "" {
		client.SendJSON(ws.Message{
			Type:    "error",
			Payload: map[string]interface{}{"message": "pending_id is required"},
		})
		return "", false
	}
	return data.PendingID, true
}

// handlePing responds with a pong so the client can measure latency.
//
// Client sends:  { "type": "ping", "payload": {} }
// Server responds: { "type": "pong", "payload": { "server_time": "..." } }
func (h *WSHandler) handlePing(client *ws.Client, payload json.RawMessage) {
	client.SendJSON(ws.Message{
		Type: "pong",
		Payload: map[string]interface{}{
			"server_time": "ok",
		},
	})
}
