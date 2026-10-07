package models

import "time"

type User struct {
	ID                string     `json:"id"`
	Username          string     `json:"username"`
	Email             string     `json:"email,omitempty"`
	PasswordHash      string     `json:"-"`
	DisplayName       *string    `json:"display_name"`
	AvatarURL         *string    `json:"avatar_url"`
	CountryCode       *string    `json:"country_code"`
	PreferredLanguage string     `json:"preferred_language"`
	Role              string     `json:"role"`
	Status            string     `json:"status"`
	IsGuest           bool       `json:"is_guest"` // throwaway demo account
	EmailVerifiedAt   *time.Time `json:"email_verified_at"`
	LastLoginAt       *time.Time `json:"last_login_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type UserRating struct {
	UserID           string    `json:"user_id"`
	ExamCategoryID   string    `json:"exam_category_id"`
	ExamCategoryCode string    `json:"exam_category_code,omitempty"`
	ExamCategoryName string    `json:"exam_category_name,omitempty"`
	Rating           int       `json:"rating"`
	MatchesPlayed    int       `json:"matches_played"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type UserStatistics struct {
	UserID                string  `json:"user_id"`
	ExamCategoryID        string  `json:"exam_category_id"`
	ExamCategoryCode      string  `json:"exam_category_code"`
	ExamCategoryName      string  `json:"exam_category_name"`
	TotalMatches          int     `json:"total_matches"`
	Wins                  int     `json:"wins"`
	Losses                int     `json:"losses"`
	Draws                 int     `json:"draws"`
	CurrentWinStreak      int     `json:"current_win_streak"`
	LongestWinStreak      int     `json:"longest_win_streak"`
	LongestLosingStreak   int     `json:"longest_losing_streak"`
	TotalQuestionsSolved  int     `json:"total_questions_solved"`
	TotalPracticeSessions int     `json:"total_practice_sessions"`
	OverallAccuracy       float64 `json:"overall_accuracy"`
	AvgSolvingTimeMs      *int    `json:"avg_solving_time_ms"`
}

type LeaderboardEntry struct {
	Rank          int     `json:"rank"`
	UserID        string  `json:"user_id"`
	Username      string  `json:"username"`
	DisplayName   *string `json:"display_name"`
	AvatarURL     *string `json:"avatar_url"`
	Rating        int     `json:"rating"`
	MatchesPlayed int     `json:"matches_played"`
}
