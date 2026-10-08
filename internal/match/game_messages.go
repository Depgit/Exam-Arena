package match

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/exam-arena/internal/platform/realtime"
)

// GameMessages handles the WebSocket messages a player sends during
// matchmaking and a match: submitting an answer, accepting or declining a
// proposed match.
type GameMessages struct {
	hub          *realtime.Hub
	matchService *Service
}

// RegisterGameMessages connects the match message types to the hub.
func RegisterGameMessages(hub *realtime.Hub, matchService *Service) *GameMessages {
	h := &GameMessages{
		hub:          hub,
		matchService: matchService,
	}
	h.registerHandlers()
	return h
}

// registerHandlers maps every incoming WebSocket message type to the
// function that processes it. Add new types here as the game grows.
func (h *GameMessages) registerHandlers() {
	h.hub.RegisterHandler("submit_answer", h.handleSubmitAnswer)
	h.hub.RegisterHandler("accept_match", h.handleAcceptMatch)
	h.hub.RegisterHandler("decline_match", h.handleDeclineMatch)
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
func (h *GameMessages) handleSubmitAnswer(client *realtime.Client, payload json.RawMessage) {
	var data struct {
		MatchID     string `json:"match_id"`
		QuestionID  string `json:"question_id"`
		OptionID    string `json:"option_id"`
		TimeTakenMs int    `json:"time_taken_ms"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		client.SendJSON(realtime.Message{
			Type:    "error",
			Payload: map[string]interface{}{"message": "invalid answer payload"},
		})
		return
	}

	// Validate required fields.
	if data.MatchID == "" || data.QuestionID == "" {
		client.SendJSON(realtime.Message{
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

	req := AnswerRequest{
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
		client.SendJSON(realtime.Message{
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
func (h *GameMessages) handleAcceptMatch(client *realtime.Client, payload json.RawMessage) {
	pendingID, ok := decodePendingID(client, payload)
	if !ok {
		return
	}

	if err := h.matchService.AcceptMatch(client.UserID(), pendingID); err != nil {
		slog.Warn("match accept failed",
			"user_id", client.UserID(), "pending_id", pendingID, "error", err)
		client.SendJSON(realtime.Message{
			Type:    "error",
			Payload: map[string]interface{}{"message": err.Error()},
		})
	}
}

// handleDeclineMatch turns down a proposed match. The opponent is put back
// into the matchmaking queue; the decliner is not.
func (h *GameMessages) handleDeclineMatch(client *realtime.Client, payload json.RawMessage) {
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
func decodePendingID(client *realtime.Client, payload json.RawMessage) (string, bool) {
	var data struct {
		PendingID string `json:"pending_id"`
	}
	if err := json.Unmarshal(payload, &data); err != nil || data.PendingID == "" {
		client.SendJSON(realtime.Message{
			Type:    "error",
			Payload: map[string]interface{}{"message": "pending_id is required"},
		})
		return "", false
	}
	return data.PendingID, true
}
