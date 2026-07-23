package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const baseURL = "http://localhost:8080/api/v1"
const wsURL = "ws://localhost:8080/ws"

type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error"`
}

type AuthData struct {
	Token string `json:"token"`
	User  struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"user"`
}

type WSMessage struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	// ── Register / Login two players ────────────────────────────────
	p1 := registerOrLogin("ws_player1", "ws1@test.com", "password123")
	p2 := registerOrLogin("ws_player2", "ws2@test.com", "password123")

	log.Printf("Player 1: %s (%s)", p1.User.Username, p1.User.ID)
	log.Printf("Player 2: %s (%s)", p2.User.Username, p2.User.ID)

	// Get SSC category ID
	sscID := getSSCCategoryID()
	log.Printf("SSC Category: %s", sscID)

	// ── Connect both players via WebSocket ──────────────────────────
	ws1 := connectWS(p1.Token, "P1")
	ws2 := connectWS(p2.Token, "P2")
	defer ws1.Close()
	defer ws2.Close()

	// Channel to collect match_start messages
	matchStartCh := make(chan WSMessage, 2)
	matchEndCh := make(chan WSMessage, 2)

	// ── Start reading WebSocket messages for both players ───────────
	var wg sync.WaitGroup
	wg.Add(2)

	go readMessages(ws1, "P1", matchStartCh, matchEndCh, &wg)
	go readMessages(ws2, "P2", matchStartCh, matchEndCh, &wg)

	// ── Both players join the queue ─────────────────────────────────
	log.Println("Both players joining queue...")
	joinQueue(p1.Token, sscID)
	joinQueue(p2.Token, sscID)
	log.Println("Both queued. Waiting for matchmaking engine...")

	// ── Wait for match_start from both players ──────────────────────
	var matchMsg WSMessage
	select {
	case matchMsg = <-matchStartCh:
		log.Println("Match started! Got match_start message")
	case <-time.After(15 * time.Second):
		log.Fatal("TIMEOUT: No match_start received in 15 seconds")
	}

	// Wait for the second player's match_start too
	select {
	case <-matchStartCh:
	case <-time.After(5 * time.Second):
	}

	matchID, _ := matchMsg.Payload["match_id"].(string)
	log.Printf("Match ID: %s", matchID)

	// ── Extract questions ───────────────────────────────────────────
	questionsRaw, _ := json.Marshal(matchMsg.Payload["questions"])
	var questions []struct {
		ID      string `json:"id"`
		Body    string `json:"body"`
		Options []struct {
			ID         string `json:"id"`
			OptionText string `json:"option_text"`
		} `json:"options"`
	}
	json.Unmarshal(questionsRaw, &questions)

	log.Printf("Received %d questions", len(questions))
	for i, q := range questions {
		log.Printf("  Q%d: %s", i+1, truncate(q.Body, 60))
	}

	// ── Simulate both players answering ─────────────────────────────
	// Player 1: answers all questions, always picks option 1 (some correct, some not)
	// Player 2: answers all questions, always picks option 1 (same strategy)
	// This tests the full flow — score calculation, broadcasting, match end.

	log.Println("")
	log.Println("=== Player 1 answering questions ===")
	for i, q := range questions {
		if len(q.Options) == 0 {
			continue
		}
		answer := WSMessage{
			Type: "submit_answer",
			Payload: map[string]interface{}{
				"match_id":      matchID,
				"question_id":   q.ID,
				"option_id":     q.Options[0].ID, // always pick first option
				"time_taken_ms": 3000 + (i * 500),
			},
		}
		data, _ := json.Marshal(answer)
		ws1.WriteMessage(websocket.TextMessage, data)
		log.Printf("  P1 answered Q%d: %s → %s", i+1, truncate(q.Body, 40), q.Options[0].OptionText)
		time.Sleep(500 * time.Millisecond) // simulate thinking time
	}

	log.Println("")
	log.Println("=== Player 2 answering questions ===")
	for i, q := range questions {
		if len(q.Options) == 0 {
			continue
		}
		// Player 2 picks the LAST option to get different results
		lastOpt := q.Options[len(q.Options)-1]
		answer := WSMessage{
			Type: "submit_answer",
			Payload: map[string]interface{}{
				"match_id":      matchID,
				"question_id":   q.ID,
				"option_id":     lastOpt.ID,
				"time_taken_ms": 5000 + (i * 300),
			},
		}
		data, _ := json.Marshal(answer)
		ws2.WriteMessage(websocket.TextMessage, data)
		log.Printf("  P2 answered Q%d: %s → %s", i+1, truncate(q.Body, 40), lastOpt.OptionText)
		time.Sleep(300 * time.Millisecond)
	}

	// ── Wait for match_end ──────────────────────────────────────────
	log.Println("")
	log.Println("Waiting for match_end...")
	select {
	case endMsg := <-matchEndCh:
		log.Println("========================================")
		log.Println("MATCH ENDED!")
		log.Println("========================================")
		resultsJSON, _ := json.MarshalIndent(endMsg.Payload, "", "  ")
		fmt.Println(string(resultsJSON))
	case <-time.After(30 * time.Second):
		log.Println("TIMEOUT waiting for match_end (may have already ended)")
	}

	// ── Check final stats via REST ──────────────────────────────────
	time.Sleep(2 * time.Second) // wait for async DB writes
	log.Println("")
	log.Println("=== Final Player Stats ===")
	checkStats(p1.User.ID, p1.User.Username)
	checkStats(p2.User.ID, p2.User.Username)

	log.Println("")
	log.Println("=== Leaderboard ===")
	checkLeaderboard()

	log.Println("")
	log.Println("TEST COMPLETE")

	// Keep alive briefly to drain remaining WS messages
	time.Sleep(3 * time.Second)
}

// ── Helper functions ─────────────────────────────────────────────────────

func registerOrLogin(username, email, password string) AuthData {
	// Try register first
	body := fmt.Sprintf(`{"username":"%s","email":"%s","password":"%s"}`, username, email, password)
	resp := doPost(baseURL+"/auth/register", body, "")
	if resp.Success {
		var auth AuthData
		json.Unmarshal(resp.Data, &auth)
		return auth
	}

	// Already exists — login
	body = fmt.Sprintf(`{"login":"%s","password":"%s"}`, username, password)
	resp = doPost(baseURL+"/auth/login", body, "")
	if !resp.Success {
		log.Fatalf("Failed to register or login %s: %s", username, resp.Error)
	}
	var auth AuthData
	json.Unmarshal(resp.Data, &auth)
	return auth
}

func getSSCCategoryID() string {
	resp, err := http.Get(baseURL + "/../health") // We'll get it from leaderboard
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	// Query DB directly through the leaderboard endpoint
	// OR we can just hardcode a query. Let's use the match endpoint.
	// Actually simplest: use admin stats or query the DB.
	// For test script, query via psql:
	log.Println("Getting SSC category ID from leaderboard endpoint...")

	// Use the leaderboard endpoint which calls GetCategoryByCode
	r, err := http.Get(baseURL + "/leaderboard/SSC?limit=1")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()
	bodyBytes, _ := io.ReadAll(r.Body)

	// If leaderboard works, we need the category ID from the DB
	// Let's just parse the users table for it. Simpler: hardcode fetch.
	log.Printf("Leaderboard response: %s", truncate(string(bodyBytes), 200))

	// The proper way: call a categories endpoint or embed in env.
	// For this test, we'll get it via a direct DB query command.
	// Fall back: parse from error or build a /categories endpoint.

	// SIMPLEST: let's add a quick categories lookup
	// For now, we know the seed creates it. Let's do a psql:
	// Actually in a Go test we can just use the register response
	// or add a helper endpoint. Let's use a different approach:

	// Just try to get categories from the DB via the test
	// We'll use an env var or default
	id := os.Getenv("SSC_CATEGORY_ID")
	if id != "" {
		return id
	}

	// Auto-detect: try to queue and see what error we get,
	// or just query. For simplicity in this test script:
	log.Println("SSC_CATEGORY_ID not set. Trying to find it...")
	log.Println("Run: export SSC_CATEGORY_ID=$(docker exec examarena-db psql -U examarena -d examarena -t -A -c \"SELECT id FROM exam_categories WHERE code = 'SSC'\")")
	log.Fatal("Set SSC_CATEGORY_ID env var and retry")
	return ""
}

func connectWS(token, label string) *websocket.Conn {
	url := fmt.Sprintf("%s?token=%s", wsURL, token)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalf("[%s] WebSocket connect failed: %v", label, err)
	}
	log.Printf("[%s] WebSocket connected", label)

	// Read the "connected" welcome message
	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Fatalf("[%s] Failed to read welcome: %v", label, err)
	}
	log.Printf("[%s] Welcome: %s", label, truncate(string(msg), 100))

	return conn
}

func readMessages(conn *websocket.Conn, label string, matchStart, matchEnd chan WSMessage, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("[%s] WebSocket closed normally", label)
			} else {
				log.Printf("[%s] WebSocket read error: %v", label, err)
			}
			return
		}

		// Handle multiple messages in one frame (newline-separated)
		parts := strings.Split(string(msg), "\n")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			var wsMsg WSMessage
			if err := json.Unmarshal([]byte(part), &wsMsg); err != nil {
				log.Printf("[%s] Failed to parse: %s", label, truncate(part, 100))
				continue
			}

			switch wsMsg.Type {
			case "match_start":
				log.Printf("[%s] *** MATCH STARTED ***", label)
				select {
				case matchStart <- wsMsg:
				default:
				}

			case "score_update":
				userID, _ := wsMsg.Payload["user_id"].(string)
				isCorrect, _ := wsMsg.Payload["is_correct"].(bool)
				points, _ := wsMsg.Payload["points_earned"].(float64)
				log.Printf("[%s] Score update: user=%s correct=%v points=%.0f",
					label, truncate(userID, 8), isCorrect, points)

			case "time_update":
				remaining, _ := wsMsg.Payload["remaining_seconds"].(float64)
				log.Printf("[%s] Time remaining: %.0fs", label, remaining)

			case "match_end":
				log.Printf("[%s] *** MATCH ENDED ***", label)
				select {
				case matchEnd <- wsMsg:
				default:
				}

			case "error":
				errMsg, _ := wsMsg.Payload["message"].(string)
				log.Printf("[%s] ERROR: %s", label, errMsg)

			default:
				log.Printf("[%s] Message type=%s", label, wsMsg.Type)
			}
		}
	}
}

func joinQueue(token, categoryID string) {
	body := fmt.Sprintf(`{"exam_category_id":"%s","match_type":"ranked"}`, categoryID)
	resp := doPost(baseURL+"/matches/queue", body, token)
	if !resp.Success {
		log.Printf("Failed to join queue: %s", resp.Error)
	}
}

func checkStats(userID, username string) {
	r, err := http.Get(baseURL + "/users/" + userID + "/stats")
	if err != nil {
		log.Printf("Failed to get stats for %s: %v", username, err)
		return
	}
	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)
	log.Printf("%s stats: %s", username, truncate(string(body), 300))
}

func checkLeaderboard() {
	r, err := http.Get(baseURL + "/leaderboard/SSC?limit=10")
	if err != nil {
		log.Printf("Failed to get leaderboard: %v", err)
		return
	}
	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)
	log.Printf("Leaderboard: %s", truncate(string(body), 500))
}

func doPost(url, body, token string) APIResponse {
	req, _ := http.NewRequest("POST", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("HTTP error: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var apiResp APIResponse
	json.Unmarshal(respBody, &apiResp)
	return apiResp
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
