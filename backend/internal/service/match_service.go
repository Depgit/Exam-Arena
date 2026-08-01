package service

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"sort"
	"time"

	appCache "github.com/exam-arena/internal/cache"
	"github.com/exam-arena/internal/matchmaking"
	"github.com/exam-arena/internal/models"
	"github.com/exam-arena/internal/repository"
	"github.com/exam-arena/internal/utils"
	"github.com/exam-arena/internal/ws"
)

// ── Constants ────────────────────────────────────────────────────────

const (
	QuestionsPerMatch = 10
	MatchTimerSeconds = 120
	PointsPerCorrect  = 100
	TimeBonusFull     = 50 // bonus if answered in < 10s
	TimeBonusHalf     = 25 // bonus if answered in < 30s
)

// ── Live match state (in memory only) ────────────────────────────────
// This struct lives in the QuestionBank cache for the duration of the match.
// It is NEVER written to Postgres — only the final results are persisted.

type LiveMatch struct {
	MatchID      string
	CategoryID   string
	MatchType    string
	Questions    []models.Question      // full objects including correct answers
	PlayerStates map[string]*LivePlayer // userID → state
	StartedAt    time.Time
	TimerSeconds int
	cancelTimer  context.CancelFunc
}

type LivePlayer struct {
	UserID      string
	Username    string
	Score       int
	AnsweredIDs map[string]bool // questionID → answered (bool = correct)
}

func (lm *LiveMatch) allAnswered() bool {
	for _, p := range lm.PlayerStates {
		if len(p.AnsweredIDs) < len(lm.Questions) {
			return false
		}
	}
	return true
}

func (lm *LiveMatch) scoreBoard() []map[string]interface{} {
	var board []map[string]interface{}
	for _, p := range lm.PlayerStates {
		correct := 0
		for _, isCorrect := range p.AnsweredIDs {
			if isCorrect {
				correct++
			}
		}
		board = append(board, map[string]interface{}{
			"user_id":            p.UserID,
			"username":           p.Username,
			"score":              p.Score,
			"questions_answered": len(p.AnsweredIDs),
			"correct":            correct,
		})
	}
	sort.Slice(board, func(i, j int) bool {
		return board[i]["score"].(int) > board[j]["score"].(int)
	})
	return board
}

// ── Service ───────────────────────────────────────────────────────────

type MatchService struct {
	matchRepo    *repository.MatchRepo
	userRepo     *repository.UserRepo
	hub          *ws.Hub
	cache        appCache.Cache         // general KV cache  (match states go here)
	questionBank *appCache.QuestionBank // pre-warmed questions
}

func NewMatchService(
	matchRepo *repository.MatchRepo,
	userRepo *repository.UserRepo,
	hub *ws.Hub,
	cache appCache.Cache,
	questionBank *appCache.QuestionBank,
) *MatchService {
	return &MatchService{
		matchRepo:    matchRepo,
		userRepo:     userRepo,
		hub:          hub,
		cache:        cache,
		questionBank: questionBank,
	}
}

// ── Entry point called by matchmaking.Engine ─────────────────────────

// StartMatchForPair is registered as the Engine's OnMatchFound callback.
// It runs in its own goroutine (the Engine spawns one per pair).
//
// Flow:
//  1. Get questions from QuestionBank (in-memory, instant)
//  2. Create DB record (persists the match skeleton)
//  3. Add players to DB
//  4. Build LiveMatch and store in cache
//  5. Push match_start to both players via WebSocket
//  6. Start countdown timer goroutine
func (s *MatchService) StartMatchForPair(ctx context.Context, pair matchmaking.PairedPlayers) {
	slog.Info("starting match",
		"player_a", pair.PlayerA.Username,
		"player_b", pair.PlayerB.Username,
		"category", pair.ExamCategoryID,
	)

	// ── Step 1: Get questions from in-memory bank (zero DB calls) ───
	questions, err := s.questionBank.GetRandom(pair.ExamCategoryID, QuestionsPerMatch)
	if err != nil {
		slog.Error("no questions available", "error", err, "category", pair.ExamCategoryID)
		s.notifyMatchFailed(pair, "not enough questions available for this category")
		return
	}

	// ── Step 2: Persist match skeleton to Postgres ──────────────────
	match, err := s.matchRepo.CreateMatch(
		ctx,
		pair.MatchType,
		pair.ExamCategoryID,
		MatchTimerSeconds,
		2,
		nil, // no room code for ranked
	)
	if err != nil {
		slog.Error("failed to create match in DB", "error", err)
		s.notifyMatchFailed(pair, "server error, please try again")
		return
	}

	// Persist which questions were used (for replay / audit)
	qIDs := make([]string, len(questions))
	for i, q := range questions {
		qIDs[i] = q.ID
	}
	if err := s.matchRepo.AddMatchQuestions(ctx, match.ID, qIDs); err != nil {
		slog.Error("failed to save match questions", "error", err)
	}

	// ── Step 3: Add players to DB ────────────────────────────────────
	players := []matchmaking.Entry{pair.PlayerA, pair.PlayerB}
	for _, p := range players {
		ratingBefore := p.Rating
		if _, err := s.matchRepo.AddPlayer(ctx, match.ID, p.UserID, &ratingBefore); err != nil {
			slog.Error("failed to add player to match", "error", err, "user", p.UserID)
		}
	}

	// Mark match as in_progress
	if err := s.matchRepo.UpdateMatchStatus(ctx, match.ID, "in_progress"); err != nil {
		slog.Error("failed to start match", "error", err)
	}

	// ── Step 4: Build in-memory LiveMatch ────────────────────────────
	timerCtx, cancelTimer := context.WithCancel(context.Background())

	liveMatch := &LiveMatch{
		MatchID:      match.ID,
		CategoryID:   pair.ExamCategoryID,
		MatchType:    pair.MatchType,
		Questions:    questions,
		PlayerStates: make(map[string]*LivePlayer),
		StartedAt:    time.Now(),
		TimerSeconds: MatchTimerSeconds,
		cancelTimer:  cancelTimer,
	}
	for _, p := range players {
		liveMatch.PlayerStates[p.UserID] = &LivePlayer{
			UserID:      p.UserID,
			Username:    p.Username,
			Score:       0,
			AnsweredIDs: make(map[string]bool),
		}
	}

	// Store in cache — key: "live_match:<matchID>"
	s.cache.Set(liveMatchKey(match.ID), liveMatch, 35*time.Minute)

	// ── Step 5: Build player-safe question list (no correct answers) ─
	playerQuestions := toPlayerQuestions(questions)

	// ── Step 6: Send match_start to both players ─────────────────────
	playerInfos := []map[string]interface{}{
		{"user_id": pair.PlayerA.UserID, "username": pair.PlayerA.Username, "rating": pair.PlayerA.Rating},
		{"user_id": pair.PlayerB.UserID, "username": pair.PlayerB.Username, "rating": pair.PlayerB.Rating},
	}

	startMsg := ws.Message{
		Type: "match_start",
		Payload: map[string]interface{}{
			"match_id":      match.ID,
			"match_type":    pair.MatchType,
			"timer_seconds": MatchTimerSeconds,
			"questions":     playerQuestions,
			"players":       playerInfos,
		},
	}
	for _, p := range players {
		s.hub.SendToUser(p.UserID, startMsg)
	}

	// ── Step 7: Start countdown timer ────────────────────────────────
	go s.runMatchTimer(timerCtx, match.ID, players, time.Duration(MatchTimerSeconds)*time.Second)
}

// ── Answer submission (called from WebSocket handler) ─────────────────

type AnswerRequest struct {
	MatchID     string
	UserID      string
	QuestionID  string
	OptionID    string // empty string means "time out / skipped"
	TimeTakenMs int
}

func (s *MatchService) SubmitAnswer(ctx context.Context, req AnswerRequest) error {
	// ── Load live match from cache ────────────────────────────────────
	raw, ok := s.cache.Get(liveMatchKey(req.MatchID))
	if !ok {
		return fmt.Errorf("match %s not found or already ended", req.MatchID)
	}
	lm := raw.(*LiveMatch)

	player, exists := lm.PlayerStates[req.UserID]
	if !exists {
		return fmt.Errorf("user %s is not a player in match %s", req.UserID, req.MatchID)
	}

	// Idempotent: ignore duplicate answers for the same question
	if _, alreadyAnswered := player.AnsweredIDs[req.QuestionID]; alreadyAnswered {
		return nil
	}

	// ── Find the question and check correctness in memory ────────────
	// No DB call — the correct answer is already in the LiveMatch struct.
	isCorrect := false
	if req.OptionID != "" {
		isCorrect = s.isCorrectOption(lm.Questions, req.QuestionID, req.OptionID)
	}

	// ── Calculate score ───────────────────────────────────────────────
	pointsEarned := 0
	if isCorrect {
		pointsEarned = PointsPerCorrect
		switch {
		case req.TimeTakenMs < 10_000:
			pointsEarned += TimeBonusFull
		case req.TimeTakenMs < 30_000:
			pointsEarned += TimeBonusHalf
		}
	}
	player.Score += pointsEarned
	player.AnsweredIDs[req.QuestionID] = isCorrect

	// ── Persist answer to Postgres asynchronously ────────────────────
	// The match result does NOT depend on this write completing — the
	// LiveMatch in cache is the source of truth until the match ends.
	go func() {
		ans := &models.MatchAnswer{
			MatchID:     req.MatchID,
			UserID:      req.UserID,
			QuestionID:  req.QuestionID,
			IsCorrect:   isCorrect,
			TimeTakenMs: req.TimeTakenMs,
		}
		if req.OptionID != "" {
			ans.SelectedOptionID = &req.OptionID
		}
		if err := s.matchRepo.SaveAnswer(context.Background(), ans); err != nil {
			slog.Error("failed to persist answer", "error", err)
		}
		if err := s.matchRepo.UpdatePlayerScore(context.Background(), req.MatchID, req.UserID, player.Score); err != nil {
			slog.Error("failed to update player score", "error", err)
		}
	}()

	// ── Broadcast live score update to all players ────────────────────
	s.broadcastScoreUpdate(lm, req.UserID, req.QuestionID, isCorrect, pointsEarned)

	// ── Check if match is over ─────────────────────────────────────────
	if lm.allAnswered() {
		lm.cancelTimer() // stop the timer goroutine
		go s.endMatch(context.Background(), lm)
	}

	return nil
}

// isCorrectOption checks correctness from the in-memory question pool.
// This is the key insight: correct answers were loaded into LiveMatch
// at match creation so we NEVER call the DB during gameplay.
func (s *MatchService) isCorrectOption(questions []models.Question, questionID, optionID string) bool {
	for _, q := range questions {
		if q.ID != questionID {
			continue
		}
		for _, o := range q.Options {
			if o.ID == optionID {
				return o.IsCorrect
			}
		}
	}
	return false
}

// ── Timer ─────────────────────────────────────────────────────────────

func (s *MatchService) runMatchTimer(
	ctx context.Context,
	matchID string,
	players []matchmaking.Entry,
	duration time.Duration,
) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	deadline := time.NewTimer(duration)
	defer deadline.Stop()

	start := time.Now()

	for {
		select {
		case <-ctx.Done():
			// cancelTimer() was called — match ended naturally
			return

		case <-ticker.C:
			remaining := duration - time.Since(start)
			if remaining < 0 {
				remaining = 0
			}
			msg := ws.Message{
				Type: "time_update",
				Payload: map[string]interface{}{
					"match_id":          matchID,
					"remaining_seconds": int(remaining.Seconds()),
				},
			}
			for _, p := range players {
				s.hub.SendToUser(p.UserID, msg)
			}

		case <-deadline.C:
			// Time's up — force end the match
			slog.Info("match timer expired", "match_id", matchID)
			raw, ok := s.cache.Get(liveMatchKey(matchID))
			if !ok {
				return // already ended
			}
			go s.endMatch(context.Background(), raw.(*LiveMatch))
			return
		}
	}
}

// ── Match end ──────────────────────────────────────────────────────────

func (s *MatchService) endMatch(ctx context.Context, lm *LiveMatch) {
	// Guard: remove from cache immediately so concurrent calls (timer +
	// allAnswered) can't double-end the same match.
	if !s.cache.DeleteIfPresent(liveMatchKey(lm.MatchID)) {
		return // another goroutine already ended it
	}

	if err := s.matchRepo.UpdateMatchStatus(ctx, lm.MatchID, "completed"); err != nil {
		slog.Error("failed to mark match completed", "error", err)
	}

	// ── Build ranked results ──────────────────────────────────────────
	type result struct {
		player  *LivePlayer
		correct int
		total   int
	}
	var results []result
	for _, p := range lm.PlayerStates {
		correct := 0
		for _, isCorrect := range p.AnsweredIDs {
			if isCorrect {
				correct++
			}
		}
		results = append(results, result{
			player:  p,
			correct: correct,
			total:   len(lm.Questions),
		})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].player.Score > results[j].player.Score
	})

	// ── Elo calculation (2-player only) ──────────────────────────────
	if len(results) == 2 {
		a, b := results[0], results[1]

		ratingA, _ := s.userRepo.GetRating(ctx, a.player.UserID, lm.CategoryID)
		ratingB, _ := s.userRepo.GetRating(ctx, b.player.UserID, lm.CategoryID)

		if ratingA != nil && ratingB != nil {
			var scoreA float64
			switch {
			case a.player.Score > b.player.Score:
				scoreA = 1.0
			case a.player.Score == b.player.Score:
				scoreA = 0.5
			default:
				scoreA = 0.0
			}

			newRA, newRB := utils.CalculateElo(ratingA.Rating, ratingB.Rating, scoreA)
			deltaA := newRA - ratingA.Rating
			deltaB := newRB - ratingB.Rating

			// Persist asynchronously — final results were already sent to players
			go func() {
				bgCtx := context.Background()
				s.userRepo.UpdateRating(bgCtx, a.player.UserID, lm.CategoryID, newRA, ratingA.MatchesPlayed+1)
				s.userRepo.UpdateRating(bgCtx, b.player.UserID, lm.CategoryID, newRB, ratingB.MatchesPlayed+1)
				s.userRepo.InsertRatingHistory(bgCtx, a.player.UserID, lm.CategoryID, lm.MatchID, ratingA.Rating, newRA, deltaA)
				s.userRepo.InsertRatingHistory(bgCtx, b.player.UserID, lm.CategoryID, lm.MatchID, ratingB.Rating, newRB, deltaB)
				s.matchRepo.UpdatePlayerRating(bgCtx, lm.MatchID, a.player.UserID, newRA, deltaA, 1)
				s.matchRepo.UpdatePlayerRating(bgCtx, lm.MatchID, b.player.UserID, newRB, deltaB, 2)
				s.userRepo.UpdateStatistics(bgCtx, a.player.UserID, lm.CategoryID, scoreA >= 0.5, a.correct, a.total)
				s.userRepo.UpdateStatistics(bgCtx, b.player.UserID, lm.CategoryID, scoreA < 0.5, b.correct, b.total)
				s.cache.DeletePrefix("leaderboard:")
			}()

			// Attach deltas to results for the WS message
			_ = deltaA
			_ = deltaB
		}
	}

	// ── Send match_end to all players ────────────────────────────────
	playerResults := make([]map[string]interface{}, len(results))
	for i, r := range results {
		playerResults[i] = map[string]interface{}{
			"user_id":  r.player.UserID,
			"username": r.player.Username,
			"score":    r.player.Score,
			"rank":     i + 1,
			"correct":  r.correct,
			"total":    r.total,
		}
	}

	endMsg := ws.Message{
		Type: "match_end",
		Payload: map[string]interface{}{
			"match_id": lm.MatchID,
			"results":  playerResults,
		},
	}
	for _, p := range lm.PlayerStates {
		s.hub.SendToUser(p.UserID, endMsg)
	}

	slog.Info("match ended",
		"match_id", lm.MatchID,
		"winner", results[0].player.Username,
		"score_a", results[0].player.Score,
		"score_b", results[1].player.Score,
	)
}

// ── Helpers ───────────────────────────────────────────────────────────

func (s *MatchService) broadcastScoreUpdate(lm *LiveMatch, answererID, questionID string, isCorrect bool, pointsEarned int) {
	msg := ws.Message{
		Type: "score_update",
		Payload: map[string]interface{}{
			"match_id":      lm.MatchID,
			"user_id":       answererID,
			"question_id":   questionID,
			"is_correct":    isCorrect,
			"points_earned": pointsEarned,
			"scoreboard":    lm.scoreBoard(),
		},
	}
	for _, p := range lm.PlayerStates {
		s.hub.SendToUser(p.UserID, msg)
	}
}

func (s *MatchService) notifyMatchFailed(pair matchmaking.PairedPlayers, reason string) {
	msg := ws.Message{
		Type:    "match_failed",
		Payload: map[string]interface{}{"reason": reason},
	}
	s.hub.SendToUser(pair.PlayerA.UserID, msg)
	s.hub.SendToUser(pair.PlayerB.UserID, msg)
}

func liveMatchKey(matchID string) string {
	return "live_match:" + matchID
}

func toPlayerQuestions(questions []models.Question) []models.QuestionForPlayer {
	result := make([]models.QuestionForPlayer, len(questions))
	for i, q := range questions {
		opts := make([]models.OptionForPlayer, len(q.Options))
		for j, o := range q.Options {
			opts[j] = models.OptionForPlayer{
				ID:         o.ID,
				OptionText: o.OptionText,
				OrderIndex: o.OrderIndex,
			}
		}
		result[i] = models.QuestionForPlayer{
			ID:                   q.ID,
			QuestionType:         q.QuestionType,
			Difficulty:           q.Difficulty,
			Body:                 q.Body,
			EstimatedTimeSeconds: q.EstimatedTimeSeconds,
			Options:              opts,
			OrderIndex:           i + 1,
		}
	}
	return result
}

func generateRoomCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

// ── Friend Match ──────────────────────────────────────────────────────────
//
// Flow:
//   Player A calls CreateFriendMatch → gets a room_code back
//   Player A shares the code out-of-band (chat, link, etc.)
//   Player B calls JoinFriendMatch(room_code)
//   When the second player joins, the match starts automatically.

// CreateFriendMatch creates a waiting match with a random room code and
// adds the creator as the first player. The match stays in "waiting"
// status until a second player joins via JoinFriendMatch.
func (s *MatchService) CreateFriendMatch(ctx context.Context, userID, username, categoryID string) (*models.Match, string, error) {
	roomCode := generateRoomCode()

	match, err := s.matchRepo.CreateMatch(ctx, "friend", categoryID, MatchTimerSeconds, 2, &roomCode)
	if err != nil {
		return nil, "", fmt.Errorf("create friend match: %w", err)
	}

	// Ensure rating row exists so we can record rating_before.
	rating, err := s.userRepo.EnsureRating(ctx, userID, categoryID)
	if err != nil {
		return nil, "", fmt.Errorf("ensure rating: %w", err)
	}

	ratingBefore := rating.Rating
	if _, err := s.matchRepo.AddPlayer(ctx, match.ID, userID, &ratingBefore); err != nil {
		return nil, "", fmt.Errorf("add creator to match: %w", err)
	}

	slog.Info("friend match created",
		"match_id", match.ID,
		"room_code", roomCode,
		"creator", username,
	)
	return match, roomCode, nil
}

// JoinFriendMatch adds a second player to an existing friend match room.
// When the room reaches max_players the match starts immediately.
func (s *MatchService) JoinFriendMatch(ctx context.Context, userID, username, roomCode string) (*models.Match, error) {
	// Look up the waiting match by room code.
	match, err := s.matchRepo.GetMatchByRoomCode(ctx, roomCode)
	if err != nil {
		return nil, fmt.Errorf("lookup room code: %w", err)
	}
	if match == nil {
		return nil, fmt.Errorf("room %q not found or already started", roomCode)
	}

	// Guard: don't let the creator join their own room as the second player.
	currentPlayers, err := s.matchRepo.GetMatchPlayers(ctx, match.ID)
	if err != nil {
		return nil, fmt.Errorf("get current players: %w", err)
	}
	for _, p := range currentPlayers {
		if p.UserID == userID {
			return nil, fmt.Errorf("you are already in this room")
		}
	}

	// Guard: room must not be full.
	if len(currentPlayers) >= match.MaxPlayers {
		return nil, fmt.Errorf("room is full")
	}

	// Add the joiner.
	rating, err := s.userRepo.EnsureRating(ctx, userID, match.ExamCategoryID)
	if err != nil {
		return nil, fmt.Errorf("ensure rating: %w", err)
	}
	ratingBefore := rating.Rating
	if _, err := s.matchRepo.AddPlayer(ctx, match.ID, userID, &ratingBefore); err != nil {
		return nil, fmt.Errorf("add joiner to match: %w", err)
	}

	slog.Info("player joined friend match",
		"match_id", match.ID,
		"room_code", roomCode,
		"user", username,
	)

	// If the room is now full, kick off the match in a goroutine so this
	// HTTP handler returns immediately.
	newPlayerCount := len(currentPlayers) + 1
	if newPlayerCount >= match.MaxPlayers {
		// Collect all player entries for the engine callback.
		allPlayers, err := s.matchRepo.GetMatchPlayers(ctx, match.ID)
		if err != nil {
			return nil, fmt.Errorf("collect players for start: %w", err)
		}

		go s.startFriendMatchGame(match, allPlayers)
	}

	return match, nil
}

// startFriendMatchGame is called in a goroutine once a friend room is full.
// It mirrors what StartMatchForPair does for ranked games.
func (s *MatchService) startFriendMatchGame(match *models.Match, dbPlayers []models.MatchPlayer) {
	ctx := context.Background()

	// Fetch questions from the in-memory bank.
	questions, err := s.questionBank.GetRandom(match.ExamCategoryID, QuestionsPerMatch)
	if err != nil {
		slog.Error("not enough questions for friend match",
			"error", err,
			"category", match.ExamCategoryID,
		)
		for _, p := range dbPlayers {
			s.hub.SendToUser(p.UserID, ws.Message{
				Type:    "match_failed",
				Payload: map[string]interface{}{"reason": "not enough questions for this category"},
			})
		}
		return
	}

	// Persist the question list.
	qIDs := make([]string, len(questions))
	for i, q := range questions {
		qIDs[i] = q.ID
	}
	if err := s.matchRepo.AddMatchQuestions(ctx, match.ID, qIDs); err != nil {
		slog.Error("failed to save friend match questions", "error", err)
	}

	// Transition status.
	if err := s.matchRepo.UpdateMatchStatus(ctx, match.ID, "in_progress"); err != nil {
		slog.Error("failed to start friend match", "error", err)
	}

	// Build in-memory LiveMatch.
	timerCtx, cancelTimer := context.WithCancel(context.Background())

	liveMatch := &LiveMatch{
		MatchID:      match.ID,
		CategoryID:   match.ExamCategoryID,
		MatchType:    "friend",
		Questions:    questions,
		PlayerStates: make(map[string]*LivePlayer),
		StartedAt:    time.Now(),
		TimerSeconds: MatchTimerSeconds,
		cancelTimer:  cancelTimer,
	}
	for _, p := range dbPlayers {
		liveMatch.PlayerStates[p.UserID] = &LivePlayer{
			UserID:      p.UserID,
			Username:    p.Username,
			Score:       0,
			AnsweredIDs: make(map[string]bool),
		}
	}
	s.cache.Set(liveMatchKey(match.ID), liveMatch, 35*time.Minute)

	// Build player-safe questions.
	playerQuestions := toPlayerQuestions(questions)

	// Build player info list.
	playerInfos := make([]map[string]interface{}, len(dbPlayers))
	for i, p := range dbPlayers {
		playerInfos[i] = map[string]interface{}{
			"user_id":  p.UserID,
			"username": p.Username,
		}
	}

	// Push match_start to every player.
	startMsg := ws.Message{
		Type: "match_start",
		Payload: map[string]interface{}{
			"match_id":      match.ID,
			"match_type":    "friend",
			"timer_seconds": MatchTimerSeconds,
			"questions":     playerQuestions,
			"players":       playerInfos,
		},
	}
	for _, p := range dbPlayers {
		s.hub.SendToUser(p.UserID, startMsg)
	}

	// Collect matchmaking.Entry-style structs for the timer (just needs UserID).
	entries := make([]matchmaking.Entry, len(dbPlayers))
	for i, p := range dbPlayers {
		entries[i] = matchmaking.Entry{UserID: p.UserID, Username: p.Username}
	}
	go s.runMatchTimer(timerCtx, match.ID, entries, time.Duration(MatchTimerSeconds)*time.Second)

	slog.Info("friend match started",
		"match_id", match.ID,
		"players", len(dbPlayers),
	)
}

// ── Match Detail Query ────────────────────────────────────────────────────

// MatchDetails is the response shape for GET /api/v1/matches/:id.
type MatchDetails struct {
	Match      *models.Match              `json:"match"`
	Players    []models.MatchPlayer       `json:"players"`
	Questions  []models.QuestionForPlayer `json:"questions,omitempty"`   // only while in_progress
	LiveScores []map[string]interface{}   `json:"live_scores,omitempty"` // only while in_progress
}

// GetMatchDetails returns the match record, its players, and — if the match
// is still live — the current scoreboard from the in-memory cache.
// Completed matches return the persisted final scores from Postgres.
func (s *MatchService) GetMatchDetails(ctx context.Context, matchID string) (*MatchDetails, error) {
	match, err := s.matchRepo.GetMatch(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("get match: %w", err)
	}
	if match == nil {
		return nil, fmt.Errorf("match %s not found", matchID)
	}

	players, err := s.matchRepo.GetMatchPlayers(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("get players: %w", err)
	}

	detail := &MatchDetails{
		Match:   match,
		Players: players,
	}

	// If the match is currently live, attach real-time scores and the
	// player-safe question list from the in-memory cache.
	if match.Status == "in_progress" {
		if raw, ok := s.cache.Get(liveMatchKey(matchID)); ok {
			lm := raw.(*LiveMatch)
			detail.LiveScores = lm.scoreBoard()
			detail.Questions = toPlayerQuestions(lm.Questions)
		}
	}

	return detail, nil
}
