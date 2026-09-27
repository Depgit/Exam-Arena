package models

import "time"

type DailyChallenge struct {
	Date             time.Time `json:"date"`
	QuestionIDs      []string  `json:"question_ids"`
	TimeLimitSeconds int       `json:"time_limit_seconds"`
}

type DailyAttempt struct {
	ID          string              `json:"id"`
	Date        time.Time           `json:"date"`
	UserID      string              `json:"user_id"`
	StartedAt   time.Time           `json:"started_at"`
	CompletedAt *time.Time          `json:"completed_at"`
	Correct     int                 `json:"correct"`
	Total       int                 `json:"total"`
	TimeTakenMs *int                `json:"time_taken_ms"`
	Answers     []DailyAnswerRecord `json:"-"`
}

// DailyAnswerRecord is one graded answer, stored as JSON on the attempt.
type DailyAnswerRecord struct {
	QuestionID string  `json:"question_id"`
	OptionID   *string `json:"option_id"`
	IsCorrect  bool    `json:"is_correct"`
}

type DailyEntry struct {
	Rank        int     `json:"rank"`
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
	Correct     int     `json:"correct"`
	Total       int     `json:"total"`
	TimeTakenMs int     `json:"time_taken_ms"`
}
