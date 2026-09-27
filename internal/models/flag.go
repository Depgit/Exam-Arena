package models

import "time"

// QuestionFlag is one player's report that a question is faulty
// (a row in reports with target_type = 'question').
type QuestionFlag struct {
	ID          string     `json:"id"`
	QuestionID  string     `json:"question_id"`
	ReporterID  string     `json:"reporter_id"`
	Reporter    string     `json:"reporter_username,omitempty"`
	Reason      string     `json:"reason"`
	Description *string    `json:"description"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`
}

// FlaggedQuestion groups the flags on one question for admin review.
type FlaggedQuestion struct {
	QuestionID     string           `json:"question_id"`
	Body           string           `json:"body"`
	QuestionStatus string           `json:"question_status"`
	Category       string           `json:"exam_category_name"`
	Topic          string           `json:"topic_name"`
	Explanation    *string          `json:"explanation"`
	Options        []QuestionOption `json:"options"`
	FlagCount      int              `json:"flag_count"`
	Reasons        map[string]int   `json:"reasons"`
	LatestFlagAt   time.Time        `json:"latest_flag_at"`
	Flags          []QuestionFlag   `json:"flags"`
}
