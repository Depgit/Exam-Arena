package models

import "time"

type Question struct {
	ID                   string           `json:"id"`
	ExamCategoryID       string           `json:"exam_category_id"`
	TopicID              string           `json:"topic_id"`
	QuestionType         string           `json:"question_type"`
	Difficulty           string           `json:"difficulty"`
	Language             string           `json:"language"`
	Body                 string           `json:"body"`
	Explanation          *string          `json:"explanation,omitempty"`
	EstimatedTimeSeconds int              `json:"estimated_time_seconds"`
	Status               string           `json:"status"`
	Options              []QuestionOption `json:"options"`
	CreatedAt            time.Time        `json:"created_at"`
}

type QuestionOption struct {
	ID         string `json:"id"`
	QuestionID string `json:"question_id"`
	OptionText string `json:"option_text"`
	IsCorrect  bool   `json:"is_correct,omitempty"` // hidden from players during match
	OrderIndex int    `json:"order_index"`
}

// QuestionForPlayer omits correct-answer information
type QuestionForPlayer struct {
	ID                   string            `json:"id"`
	QuestionType         string            `json:"question_type"`
	Difficulty           string            `json:"difficulty"`
	Body                 string            `json:"body"`
	EstimatedTimeSeconds int               `json:"estimated_time_seconds"`
	Options              []OptionForPlayer `json:"options"`
	OrderIndex           int               `json:"order_index"`
}

type OptionForPlayer struct {
	ID         string `json:"id"`
	OptionText string `json:"option_text"`
	OrderIndex int    `json:"order_index"`
}

type ExamCategory struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type Topic struct {
	ID             string  `json:"id"`
	ExamCategoryID string  `json:"exam_category_id"`
	ParentTopicID  *string `json:"parent_topic_id"`
	Name           string  `json:"name"`
}
