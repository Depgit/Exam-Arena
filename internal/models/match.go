package models

import "time"

type Match struct {
	ID             string     `json:"id"`
	MatchType      string     `json:"match_type"`
	Status         string     `json:"status"`
	ExamCategoryID string     `json:"exam_category_id"`
	RoomCode       *string    `json:"room_code,omitempty"`
	TournamentID   *string    `json:"tournament_id,omitempty"`
	TimerSeconds   int        `json:"timer_seconds"`
	MaxPlayers     int        `json:"max_players"`
	CreatedAt      time.Time  `json:"created_at"`
	StartedAt      *time.Time `json:"started_at"`
	EndedAt        *time.Time `json:"ended_at"`
}

type MatchPlayer struct {
	ID               string    `json:"id"`
	MatchID          string    `json:"match_id"`
	UserID           string    `json:"user_id"`
	Username         string    `json:"username"`
	Score            int       `json:"score"`
	FinalRank        *int      `json:"final_rank"`
	RatingBefore     *int      `json:"rating_before"`
	RatingAfter      *int      `json:"rating_after"`
	RatingDelta      *int      `json:"rating_delta"`
	ConnectionStatus string    `json:"connection_status"`
	JoinedAt         time.Time `json:"joined_at"`
}

type MatchQuestion struct {
	MatchID    string `json:"match_id"`
	QuestionID string `json:"question_id"`
	OrderIndex int    `json:"order_index"`
}

type MatchAnswer struct {
	ID               string    `json:"id"`
	MatchID          string    `json:"match_id"`
	UserID           string    `json:"user_id"`
	QuestionID       string    `json:"question_id"`
	SelectedOptionID *string   `json:"selected_option_id"`
	IsCorrect        bool      `json:"is_correct"`
	TimeTakenMs      int       `json:"time_taken_ms"`
	AnsweredAt       time.Time `json:"answered_at"`
}

type MatchResult struct {
	MatchID  string              `json:"match_id"`
	Status   string              `json:"status"`
	Players  []MatchPlayerResult `json:"players"`
	Duration int                 `json:"duration_seconds"`
}

type MatchPlayerResult struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	Score       int    `json:"score"`
	Rank        int    `json:"rank"`
	RatingDelta int    `json:"rating_delta"`
	Correct     int    `json:"correct"`
	Total       int    `json:"total"`
}

type QueueEntry struct {
	UserID         string    `json:"user_id"`
	ExamCategoryID string    `json:"exam_category_id"`
	MatchType      string    `json:"match_type"`
	Rating         int       `json:"rating"`
	QueuedAt       time.Time `json:"queued_at"`
}
