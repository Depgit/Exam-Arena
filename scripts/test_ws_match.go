package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// ── Config ──────────────────────────────────────────────────────────────

var (
	baseURL = getEnv("BASE_URL", "http://localhost:8080/api/v1")
	wsURL   = getEnv("WS_URL", "ws://localhost:8080/ws")
	sscID   = getEnv("SSC_CATEGORY_ID", "") // MUST be set via env var
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ── Types ───────────────────────────────────────────────────────────────

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
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// ── Main ────────────────────────────────────────────────────────────────

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	if sscID == "" {
		log.Fatal("SSC_CATEGORY_ID env var is required.\n" +
			"Run: export SSC_CATEGORY_ID=$(docker exec examarena-db " +
			"psql -U examarena -d examarena -t -A " +
			"-c \"SELECT id FROM exam_categories WHERE code = 'SSC'\")")
	}

	log.Printf("Using SSC Category ID: %s", sscID)

	// ── 1. Register or login two players ────────────────────────────
	log.Println("")
	log.Println("═══ STEP 1: Register/Login Players ═══")
	p1 := registerOrLogin("ws_test_p1", "wstest1@test.com", "password123")
	p2 := registerOrLogin("ws_test_p2", "wstest2@test.com", "password123")

	log.Printf("Player 1: %s (%s)", p1.User.Username, p1.User.ID)
	log.Printf("Player 2: %s (%s)", p2.User.Username, p2.User.ID)

	if p1.Token == "" || p2.Token == "" {
		log.Fatal("Token is empty — authentication failed")
	}

	// ── 2. Connect WebSocket ────────────────────────────────────────
	log.Println("")
	log.Println("═══ STEP 2: Connect WebSocket ═══")
	ws1 := connectWS(p1.Token, "P1")
	ws2 := connectWS(p2.Token, "P2")
	defer ws1.Close()
	defer ws2.Close()

	// Channels for coordination
	matchStartCh := make(chan WSMessage, 4)
	scoreCh := make(chan WSMessage, 100)
	matchEndCh := make(chan WSMessage, 4)

	// ── 3. Start message readers ────────────────────────────────────
	go readMessages(ws1, "P1", matchStartCh, scoreCh, matchEndCh)
	go readMessages(ws2, "P2", matchStartCh, scoreCh, matchEndCh)

	// ── 4. Join queue ───────────────────────────────────────────────
	log.Println("")
	log.Println("═══ STEP 3: Both Players Join Queue ═══")
	joinQueue(p1.Token, "P1")
	joinQueue(p2.Token, "P2")

	// ── 5. Wait for match_start ─────────────────────────────────────
	log.Println("")
	log.Println("═══ STEP 4: Waiting for Match ═══")

	var matchStartMsg WSMessage
	select {
	case matchStartMsg = <-matchStartCh:
		log.Println("✓ Received match_start!")
	case <-time.After(20 * time.Second):
		log.Fatal("TIMEOUT: No match_start in 20 seconds. Is the server running?")
	}

	// Drain the second player's match_start
	select {
	case <-matchStartCh:
	case <-time.After(3 * time.Second):
	}

	// ── 6. Parse match data ─────────────────────────────────────────
	var matchPayload struct {
		MatchID      string `json:"match_id"`
		MatchType    string `json:"match_type"`
		TimerSeconds int    `json:"timer_seconds"`
		Questions    []struct {
			ID      string `json:"id"`
			Body    string `json:"body"`
			Options []struct {
				ID         string `json:"id"`
				OptionText string `json:"option_text"`
			} `json:"options"`
		} `json:"questions"`
		Players []struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			Rating   int    `json:"rating"`
		} `json:"players"`
	}
	if err := json.Unmarshal(matchStartMsg.Payload, &matchPayload); err != nil {
		log.Fatalf("Failed to parse match_start: %v", err)
	}

	matchID := matchPayload.MatchID
	log.Printf("Match ID:      %s", matchID)
	log.Printf("Match Type:    %s", matchPayload.MatchType)
	log.Printf("Timer:         %d seconds", matchPayload.TimerSeconds)
	log.Printf("Questions:     %d", len(matchPayload.Questions))

	for _, p := range matchPayload.Players {
		log.Printf("  Player: %s (rating: %d)", p.Username, p.Rating)
	}

	if len(matchPayload.Questions) == 0 {
		log.Fatal("ERROR: Match started with 0 questions — question bank is empty. Did you run 'make db-seed'?")
	}

	log.Println("")
	log.Println("═══ STEP 5: Questions ═══")
	for i, q := range matchPayload.Questions {
		log.Printf("  Q%02d: %s", i+1, truncate(q.Body, 70))
		for j, opt := range q.Options {
			marker := "  "
			if j == 0 {
				marker = "→ " // P1 will pick this
			}
			log.Printf("      %s%s: %s", marker, opt.ID[:8], opt.OptionText)
		}
	}

	// ── 7. Player 1 answers all questions ───────────────────────────
	log.Println("")
	log.Println("═══ STEP 6: Player 1 Answering (picks FIRST option each time) ═══")

	for i, q := range matchPayload.Questions {
		if len(q.Options) == 0 {
			log.Printf("  P1: Q%d has no options, skipping", i+1)
			continue
		}

		chosenOpt := q.Options[0] // always pick first option
		timeTaken := 2000 + (i * 400)

		sendAnswer(ws1, matchID, q.ID, chosenOpt.ID, timeTaken)
		log.Printf("  P1 → Q%02d: chose %q (%dms)", i+1, truncate(chosenOpt.OptionText, 30), timeTaken)

		time.Sleep(300 * time.Millisecond)
	}

	// ── 8. Player 2 answers all questions ───────────────────────────
	log.Println("")
	log.Println("═══ STEP 7: Player 2 Answering (picks LAST option each time) ═══")

	for i, q := range matchPayload.Questions {
		if len(q.Options) == 0 {
			continue
		}

		chosenOpt := q.Options[len(q.Options)-1] // always pick last option
		timeTaken := 4000 + (i * 300)

		sendAnswer(ws2, matchID, q.ID, chosenOpt.ID, timeTaken)
		log.Printf("  P2 → Q%02d: chose %q (%dms)", i+1, truncate(chosenOpt.OptionText, 30), timeTaken)

		time.Sleep(200 * time.Millisecond)
	}

	// ── 9. Drain score updates ──────────────────────────────────────
	log.Println("")
	log.Println("═══ STEP 8: Draining Score Updates ═══")
	drainScores(scoreCh, 3*time.Second)

	// ── 10. Wait for match_end ──────────────────────────────────────
	log.Println("")
	log.Println("═══ STEP 9: Waiting for Match End ═══")

	select {
	case endMsg := <-matchEndCh:
		log.Println("╔══════════════════════════════════════╗")
		log.Println("║          MATCH COMPLETE!             ║")
		log.Println("╚══════════════════════════════════════╝")
		prettyPrint(endMsg.Payload)
	case <-time.After(30 * time.Second):
		log.Println("No match_end received — match may have ended before all answers")
	}

	// ── 11. Check final stats ───────────────────────────────────────
	log.Println("")
	log.Println("═══ STEP 10: Final Stats ═══")
	time.Sleep(2 * time.Second) // wait for async DB writes

	checkPlayerStats(p1.User.ID, p1.User.Username)
	checkPlayerStats(p2.User.ID, p2.User.Username)
	checkLeaderboard()

	log.Println("")
	log.Println("═══ TEST COMPLETE ═══")
	time.Sleep(2 * time.Second)
}

// ── Authentication ──────────────────────────────────────────────────────

func registerOrLogin(username, email, password string) AuthData {
	// Try register first
	body := fmt.Sprintf(`{"username":"%s","email":"%s","password":"%s"}`,
		username, email, password)
	resp := httpPost(baseURL+"/auth/register", body, "")
	if resp.Success {
		var auth AuthData
		json.Unmarshal(resp.Data, &auth)
		if auth.Token != "" {
			log.Printf("✓ Registered: %s", username)
			return auth
		}
	}

	// Already exists — login
	log.Printf("  User %s already exists, logging in...", username)
	body = fmt.Sprintf(`{"login":"%s","password":"%s"}`, username, password)
	resp = httpPost(baseURL+"/auth/login", body, "")
	if !resp.Success {
		log.Fatalf("Failed to login %s: %s", username, resp.Error)
	}
	var auth AuthData
	json.Unmarshal(resp.Data, &auth)
	if auth.Token == "" {
		log.Fatalf("Login returned empty token for %s. Response: %s", username, string(resp.Data))
	}
	log.Printf("✓ Logged in: %s", username)
	return auth
}

// ── WebSocket ───────────────────────────────────────────────────────────

func connectWS(token, label string) *websocket.Conn {
	url := fmt.Sprintf("%s?token=%s", wsURL, token)
	log.Printf("[%s] Connecting to %s", label, truncate(url, 80))

	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		// Print the HTTP response body for debugging
		if resp != nil {
			body, _ := io.ReadAll(resp.Body)
			log.Fatalf("[%s] WebSocket connect failed (HTTP %d): %s\nError: %v",
				label, resp.StatusCode, string(body), err)
		}
		log.Fatalf("[%s] WebSocket connect failed: %v", label, err)
	}
	if resp != nil {
		resp.Body.Close()
	}

	log.Printf("[%s] ✓ WebSocket connected (HTTP %d)", label, resp.StatusCode)

	// Read the "connected" welcome message
	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Fatalf("[%s] Failed to read welcome message: %v", label, err)
	}
	log.Printf("[%s] Welcome: %s", label, truncate(string(msg), 120))

	return conn
}

func readMessages(
	conn *websocket.Conn,
	label string,
	matchStartCh chan<- WSMessage,
	scoreCh chan<- WSMessage,
	matchEndCh chan<- WSMessage,
) {
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("[%s] Connection closed normally", label)
			} else {
				log.Printf("[%s] Connection error: %v", label, err)
			}
			return
		}

		// Server may batch multiple messages in one frame (newline-separated)
		for _, line := range strings.Split(string(raw), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			var msg WSMessage
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				log.Printf("[%s] parse error: %s", label, truncate(line, 80))
				continue
			}

			switch msg.Type {
			case "match_start":
				log.Printf("[%s] ═══ MATCH STARTED ═══", label)
				select {
				case matchStartCh <- msg:
				default:
				}

			case "score_update":
				var payload struct {
					UserID       string  `json:"user_id"`
					IsCorrect    bool    `json:"is_correct"`
					PointsEarned float64 `json:"points_earned"`
					Scoreboard   []struct {
						UserID   string `json:"user_id"`
						Score    int    `json:"score"`
						Answered int    `json:"questions_answered"`
					} `json:"scoreboard"`
				}
				json.Unmarshal(msg.Payload, &payload)
				log.Printf("[%s] Score: user=%s correct=%v pts=%.0f",
					label, payload.UserID[:8], payload.IsCorrect, payload.PointsEarned)
				select {
				case scoreCh <- msg:
				default:
				}

			case "time_update":
				var payload struct {
					Remaining float64 `json:"remaining_seconds"`
				}
				json.Unmarshal(msg.Payload, &payload)
				log.Printf("[%s] Timer: %.0fs remaining", label, payload.Remaining)

			case "match_end":
				log.Printf("[%s] ═══ MATCH ENDED ═══", label)
				select {
				case matchEndCh <- msg:
				default:
				}

			case "match_failed":
				var payload struct {
					Reason string `json:"reason"`
				}
				json.Unmarshal(msg.Payload, &payload)
				log.Printf("[%s] ═══ MATCH FAILED: %s ═══", label, payload.Reason)

			case "error":
				var payload struct {
					Message string `json:"message"`
				}
				json.Unmarshal(msg.Payload, &payload)
				log.Printf("[%s] ERROR from server: %s", label, payload.Message)

			case "connected", "pong":
				// already logged in connectWS
			default:
				log.Printf("[%s] Unknown message type: %s", label, msg.Type)
			}
		}
	}
}

func sendAnswer(conn *websocket.Conn, matchID, questionID, optionID string, timeTakenMs int) {
	msg := map[string]interface{}{
		"type": "submit_answer",
		"payload": map[string]interface{}{
			"match_id":      matchID,
			"question_id":   questionID,
			"option_id":     optionID,
			"time_taken_ms": timeTakenMs,
		},
	}
	data, _ := json.Marshal(msg)
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("Failed to send answer: %v", err)
	}
}

func drainScores(ch <-chan WSMessage, timeout time.Duration) {
	deadline := time.After(timeout)
	for {
		select {
		case <-ch:
			// consumed a score update
		case <-deadline:
			return
		}
	}
}

// ── HTTP helpers ────────────────────────────────────────────────────────

func httpPost(url, body, token string) APIResponse {
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(body))
	if err != nil {
		log.Fatalf("httpPost request error: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("httpPost error: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		log.Fatalf("httpPost unmarshal error: %v\nBody: %s", err, string(respBody))
	}
	return apiResp
}

func httpGet(url, token string) APIResponse {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("httpGet request error: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("httpGet error: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var apiResp APIResponse
	json.Unmarshal(respBody, &apiResp)
	return apiResp
}

// ── Reporting ───────────────────────────────────────────────────────────

func joinQueue(token, label string) {
	resp := httpPost(baseURL+"/matches/queue",
		fmt.Sprintf(`{"exam_category_id":"%s","match_type":"ranked"}`, sscID),
		token,
	)
	if resp.Success {
		log.Printf("[%s] ✓ Joined queue", label)
	} else {
		log.Printf("[%s] ✗ Queue join failed: %s", label, resp.Error)
	}
}

func checkPlayerStats(userID, username string) {
	log.Printf("--- %s ---", username)

	profileResp := httpGet(baseURL+"/users/"+userID, "")
	if profileResp.Success {
		var profile struct {
			User struct {
				Username    string `json:"username"`
				DisplayName string `json:"display_name"`
			} `json:"user"`
			Ratings []struct {
				ExamCategoryID string `json:"exam_category_id"`
				Rating         int    `json:"rating"`
				MatchesPlayed  int    `json:"matches_played"`
			} `json:"ratings"`
		}
		json.Unmarshal(profileResp.Data, &profile)
		for _, r := range profile.Ratings {
			log.Printf("  Rating: %d  Matches: %d  Category: %s",
				r.Rating, r.MatchesPlayed, r.ExamCategoryID[:8])
		}
	}

	statsResp := httpGet(baseURL+"/users/"+userID+"/stats", "")
	if statsResp.Success {
		var stats []struct {
			Wins    int `json:"wins"`
			Losses  int `json:"losses"`
			Matches int `json:"total_matches"`
		}
		json.Unmarshal(statsResp.Data, &stats)
		for _, s := range stats {
			log.Printf("  W/L: %d/%d  Total: %d", s.Wins, s.Losses, s.Matches)
		}
	}
}

func checkLeaderboard() {
	log.Println("--- Leaderboard (SSC) ---")
	resp := httpGet(baseURL+"/leaderboard/SSC?limit=5", "")
	if resp.Success {
		var entries []struct {
			Rank     int    `json:"rank"`
			Username string `json:"username"`
			Rating   int    `json:"rating"`
		}
		json.Unmarshal(resp.Data, &entries)
		for _, e := range entries {
			log.Printf("  #%d  %-15s  Rating: %d", e.Rank, e.Username, e.Rating)
		}
	}
}

func prettyPrint(data json.RawMessage) {
	var pretty bytes.Buffer
	json.Indent(&pretty, data, "", "  ")
	fmt.Println(pretty.String())
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
