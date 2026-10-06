package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/exam-arena/internal/matchmaking"
	"github.com/exam-arena/internal/ws"
)

// ── The 30-second acceptance window ───────────────────────────────────
//
// game_rules.md §4 puts an acceptance step between pairing and play:
//
//	Opponent found → 30 second acceptance window → Countdown → Battle
//
// It exists so a player who has wandered off does not drag someone else
// through a dead two-minute match. The engine does the pairing; this file
// owns the gap between "paired" and "playing".
//
// Why nothing in the matchmaking engine had to change: Engine.processPool
// already calls ClaimPair — which removes both players from the queue —
// before invoking the OnMatchFound callback. A pending pair is therefore
// already out of the pool and cannot be paired again while it waits. The
// only wiring change is pointing OnMatchFound at ProposeMatch instead of
// StartMatchForPair.

const (
	// AcceptWindowSeconds is how long each player has to accept.
	AcceptWindowSeconds = 30

	reasonDeclined     = "declined"
	reasonTimeout      = "timeout"
	reasonDisconnected = "disconnected"
)

// pendingMatch is one paired-but-not-yet-started match.
type pendingMatch struct {
	id       string
	pair     matchmaking.PairedPlayers // immutable after creation
	accepted map[string]bool           // userID → accepted; guarded by pendingRegistry.mu
	resolved bool                      // guarded by pendingRegistry.mu
	timer    *time.Timer               // the acceptance deadline
}

func (p *pendingMatch) entries() [2]matchmaking.Entry {
	return [2]matchmaking.Entry{p.pair.PlayerA, p.pair.PlayerB}
}

func (p *pendingMatch) isPlayer(userID string) bool {
	return p.pair.PlayerA.UserID == userID || p.pair.PlayerB.UserID == userID
}

func (p *pendingMatch) opponentOf(userID string) matchmaking.Entry {
	if p.pair.PlayerA.UserID == userID {
		return p.pair.PlayerB
	}
	return p.pair.PlayerA
}

// pendingRegistry tracks every match waiting on acceptance.
//
// byUser lets an inbound accept_match be resolved from the authenticated
// sender rather than trusting the pending_id in the payload, and guarantees
// a user is never in two pending matches at once.
type pendingRegistry struct {
	mu     sync.Mutex
	byID   map[string]*pendingMatch
	byUser map[string]string // userID → pending match id
}

func newPendingRegistry() *pendingRegistry {
	return &pendingRegistry{
		byID:   make(map[string]*pendingMatch),
		byUser: make(map[string]string),
	}
}

// forgetLocked removes a resolved pending match. Callers hold mu.
func (r *pendingRegistry) forgetLocked(pm *pendingMatch) {
	delete(r.byID, pm.id)
	for _, e := range pm.entries() {
		// Only clear the index if it still points at this match. A player
		// requeued from an earlier cancellation may already be in a newer
		// pending match, and that index entry must survive.
		if r.byUser[e.UserID] == pm.id {
			delete(r.byUser, e.UserID)
		}
	}
}

func newPendingID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand failing is not recoverable in any useful way here, and
		// the id is an opaque handle rather than a secret — membership is
		// always re-checked against the authenticated sender.
		return fmt.Sprintf("pm-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// ── Entry point: replaces StartMatchForPair as the engine callback ────

// ProposeMatch offers a freshly paired match to both players and waits for
// them to accept. It is registered as the Engine's OnMatchFound callback and
// runs in its own goroutine (the engine spawns one per pair).
//
// When both players accept it calls StartMatchForPair, which is unchanged —
// everything downstream of "match starts" behaves exactly as before.
func (s *MatchService) ProposeMatch(ctx context.Context, pair matchmaking.PairedPlayers) {
	pm := &pendingMatch{
		id:       newPendingID(),
		pair:     pair,
		accepted: make(map[string]bool, 2),
	}

	s.pending.mu.Lock()
	for _, e := range pm.entries() {
		if stale, ok := s.pending.byUser[e.UserID]; ok {
			// Should be unreachable: a pending player is out of the queue and
			// cannot be re-paired. Worth knowing about if it ever happens.
			slog.Warn("player paired while already pending",
				"user_id", e.UserID, "stale_pending", stale, "new_pending", pm.id)
		}
		s.pending.byUser[e.UserID] = pm.id
	}
	s.pending.byID[pm.id] = pm
	// Armed under the lock so the callback cannot observe a nil pm.timer.
	// AfterFunc runs its callback on a fresh goroutine, which will block on
	// mu until this unlocks.
	pm.timer = time.AfterFunc(AcceptWindowSeconds*time.Second, func() {
		s.resolvePending(pm.id, reasonTimeout, "")
	})
	s.pending.mu.Unlock()

	slog.Info("match proposed",
		"pending_id", pm.id,
		"player_a", pair.PlayerA.Username,
		"player_b", pair.PlayerB.Username,
		"match_type", pair.MatchType,
	)

	// Each player is told about their opponent, never about themselves in
	// the opponent slot.
	for _, me := range pm.entries() {
		opponent := pm.opponentOf(me.UserID)
		s.hub.SendToUser(me.UserID, ws.Message{
			Type: "match_found",
			Payload: map[string]interface{}{
				"pending_id":       pm.id,
				"match_type":       pair.MatchType,
				"exam_category_id": pair.ExamCategoryID,
				"accept_seconds":   AcceptWindowSeconds,
				"you": map[string]interface{}{
					"user_id":  me.UserID,
					"username": me.Username,
					"rating":   me.Rating,
				},
				"opponent": map[string]interface{}{
					"user_id":  opponent.UserID,
					"username": opponent.Username,
					"rating":   opponent.Rating,
				},
			},
		})
	}
}

// ── Accept ────────────────────────────────────────────────────────────

// AcceptMatch records one player's acceptance. Once both have accepted the
// match starts. Accepting twice is a no-op, not an error — a double tap on a
// laggy connection should not look like a failure.
func (s *MatchService) AcceptMatch(userID, pendingID string) error {
	s.pending.mu.Lock()

	pm, ok := s.pending.byID[pendingID]
	if !ok || pm.resolved {
		s.pending.mu.Unlock()
		return fmt.Errorf("this match is no longer waiting for acceptance")
	}
	if !pm.isPlayer(userID) {
		s.pending.mu.Unlock()
		return fmt.Errorf("you are not part of this match")
	}
	if pm.accepted[userID] {
		s.pending.mu.Unlock()
		return nil
	}

	pm.accepted[userID] = true
	bothAccepted := len(pm.accepted) == 2
	if bothAccepted {
		pm.resolved = true
		pm.timer.Stop()
		s.pending.forgetLocked(pm)
	}
	opponent := pm.opponentOf(userID)
	pair := pm.pair // immutable; safe to use after unlocking

	s.pending.mu.Unlock()

	if bothAccepted {
		slog.Info("match accepted by both players", "pending_id", pendingID)
		// The existing, unmodified start path. Its own goroutine because it
		// does DB work and we may be on the hub goroutine.
		go s.StartMatchForPair(context.Background(), pair)
		return nil
	}

	// Let the other player see they are the one being waited on.
	s.hub.SendToUser(opponent.UserID, ws.Message{
		Type: "opponent_accepted",
		Payload: map[string]interface{}{
			"pending_id": pendingID,
		},
	})
	return nil
}

// ── Decline, timeout, disconnect ──────────────────────────────────────

// DeclineMatch turns down a pending match. The opponent goes back in the
// queue; the decliner does not.
func (s *MatchService) DeclineMatch(userID, pendingID string) error {
	s.pending.mu.Lock()
	pm, ok := s.pending.byID[pendingID]
	if !ok || pm.resolved {
		s.pending.mu.Unlock()
		return fmt.Errorf("this match is no longer waiting for acceptance")
	}
	if !pm.isPlayer(userID) {
		s.pending.mu.Unlock()
		return fmt.Errorf("you are not part of this match")
	}
	s.pending.mu.Unlock()

	// resolvePending re-checks under the lock, so the gap here is harmless:
	// the worst case is that it has already been resolved and this is a no-op.
	s.resolvePending(pendingID, reasonDeclined, userID)
	return nil
}

// HandlePresenceChange drops a pending match when one of its players goes
// offline, so a closed tab does not make the opponent wait out the full
// window. Wired to the hub's presence hook in main.go.
func (s *MatchService) HandlePresenceChange(userID string, online bool) {
	if online {
		return
	}
	s.pending.mu.Lock()
	pendingID, ok := s.pending.byUser[userID]
	s.pending.mu.Unlock()
	if !ok {
		return
	}
	s.resolvePending(pendingID, reasonDisconnected, userID)
}

// resolvePending ends a pending match without starting it and decides, per
// player, who goes back in the queue.
//
// declinerID is the player at fault ("" for a plain timeout, where nobody is).
//
//	explicit decline / disconnect → everyone except the decliner is requeued.
//	  The other player did nothing wrong, whether or not they had tapped
//	  Accept yet.
//	timeout → only players who actually accepted are requeued. Requeuing
//	  someone who never responded would just block the next real player.
func (s *MatchService) resolvePending(pendingID, reason, declinerID string) {
	s.pending.mu.Lock()
	pm, ok := s.pending.byID[pendingID]
	if !ok || pm.resolved {
		s.pending.mu.Unlock()
		return // already started, declined or timed out
	}
	pm.resolved = true
	pm.timer.Stop()
	s.pending.forgetLocked(pm)

	accepted := make(map[string]bool, len(pm.accepted))
	for id, v := range pm.accepted {
		accepted[id] = v
	}
	s.pending.mu.Unlock()

	slog.Info("pending match resolved without starting",
		"pending_id", pendingID, "reason", reason, "by", declinerID)

	ctx := context.Background()

	for _, e := range pm.entries() {
		atFault := e.UserID == declinerID
		requeue := !atFault && (declinerID != "" || accepted[e.UserID])

		if requeue {
			// Push the ORIGINAL entry: QueuedAt is preserved, so the player
			// keeps their place and the rating tolerance their wait had
			// already earned them (see Engine.tolerance).
			//
			// Unless they have already queued for something else in the
			// meantime — Queue.Push would silently replace that newer choice.
			if !s.queue.IsQueued(e.UserID) {
				s.queue.Push(e)
			}
		} else if s.matchRepo != nil {
			// Not going back in: clear the Postgres queue mirror too.
			// ClaimPair only emptied the in-memory pool, so without this
			// RestoreFromDB would put them back after a restart and throw
			// them into a match they never asked for.
			//
			// (Nil-checked because the unit tests build a MatchService with
			// no database.)
			if err := s.matchRepo.RemoveFromQueue(ctx, e.UserID); err != nil {
				slog.Warn("failed to clear queue mirror after cancel",
					"user_id", e.UserID, "error", err)
			}
		}

		s.hub.SendToUser(e.UserID, ws.Message{
			Type: "match_cancelled",
			Payload: map[string]interface{}{
				"pending_id":        pendingID,
				"reason":            reason,
				"requeued":          requeue,
				"you_declined":      atFault,
				"opponent_username": pm.opponentOf(e.UserID).Username,
			},
		})
	}
}
