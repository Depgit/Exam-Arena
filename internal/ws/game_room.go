package ws

import (
	"sync"
)

// GameRoom tracks live game state for WebSocket broadcasts.
// Multiple game rooms run concurrently (one per active match).
type GameRoom struct {
	mu        sync.RWMutex
	MatchID   string
	PlayerIDs []string
	Scores    map[string]int
	Active    bool
}

func NewGameRoom(matchID string, playerIDs []string) *GameRoom {
	scores := make(map[string]int)
	for _, pid := range playerIDs {
		scores[pid] = 0
	}
	return &GameRoom{
		MatchID:   matchID,
		PlayerIDs: playerIDs,
		Scores:    scores,
		Active:    true,
	}
}

func (gr *GameRoom) UpdateScore(userID string, delta int) int {
	gr.mu.Lock()
	defer gr.mu.Unlock()
	gr.Scores[userID] += delta
	return gr.Scores[userID]
}

func (gr *GameRoom) GetScores() map[string]int {
	gr.mu.RLock()
	defer gr.mu.RUnlock()
	result := make(map[string]int)
	for k, v := range gr.Scores {
		result[k] = v
	}
	return result
}

func (gr *GameRoom) Close() {
	gr.mu.Lock()
	defer gr.mu.Unlock()
	gr.Active = false
}
