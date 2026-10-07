package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/exam-arena/internal/models"
	"github.com/exam-arena/internal/repository"
)

const (
	DailyQuestionCount    = 10
	DailyTimeLimitSeconds = 180
	// dailyGrace absorbs network latency on a submit sent at the deadline.
	dailyGrace = 10 * time.Second
	// dailyLeaderboardSize is how many entries the overview returns.
	dailyLeaderboardSize = 10
)

// dailyZone decides when the day rolls over: midnight IST. A fixed offset
// avoids depending on tzdata being present in the container.
var dailyZone = time.FixedZone("IST", 5*60*60+30*60)

var (
	ErrNoDailyChallenge   = errors.New("no daily challenge available today")
	ErrDailyAlreadyPlayed = errors.New("you have already played today's challenge")
	ErrDailyNotStarted    = errors.New("start today's challenge first")
)

type DailyService struct {
	dailyRepo    *repository.DailyRepo
	questionRepo *repository.QuestionRepo
	now          func() time.Time
}

func NewDailyService(dailyRepo *repository.DailyRepo, questionRepo *repository.QuestionRepo) *DailyService {
	return &DailyService{dailyRepo: dailyRepo, questionRepo: questionRepo, now: time.Now}
}

// challengeDate is today's date in dailyZone, as midnight UTC (the form
// stored in the DATE column).
func challengeDate(now time.Time) time.Time {
	y, m, d := now.In(dailyZone).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// nextReset is the next midnight in dailyZone.
func nextReset(now time.Time) time.Time {
	y, m, d := now.In(dailyZone).Date()
	return time.Date(y, m, d+1, 0, 0, 0, 0, dailyZone)
}

// ensureChallenge returns today's challenge, fixing its question set on
// first use.
func (s *DailyService) ensureChallenge(ctx context.Context, date time.Time) (*models.DailyChallenge, error) {
	c, err := s.dailyRepo.GetChallenge(ctx, date)
	if err != nil || c != nil {
		return c, err
	}
	ids, err := s.questionRepo.PickDailyQuestionIDs(ctx, date.Format("2006-01-02"), DailyQuestionCount)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, ErrNoDailyChallenge
	}
	if err := s.dailyRepo.CreateChallenge(ctx, date, ids, DailyTimeLimitSeconds); err != nil {
		return nil, err
	}
	return s.dailyRepo.GetChallenge(ctx, date)
}

func deadlineOf(a *models.DailyAttempt, c *models.DailyChallenge) time.Time {
	return a.StartedAt.Add(time.Duration(c.TimeLimitSeconds) * time.Second)
}

// expireIfOverdue closes an attempt whose time ran out without a submit,
// scoring it zero. It reports whether the attempt is now completed.
func (s *DailyService) expireIfOverdue(ctx context.Context, a *models.DailyAttempt, c *models.DailyChallenge) (bool, error) {
	if a.CompletedAt != nil {
		return true, nil
	}
	if s.now().Before(deadlineOf(a, c).Add(dailyGrace)) {
		return false, nil
	}
	limitMs := c.TimeLimitSeconds * 1000
	if _, err := s.dailyRepo.CompleteAttempt(ctx, a.ID, 0, len(c.QuestionIDs), limitMs, []models.DailyAnswerRecord{}); err != nil {
		return false, err
	}
	now := s.now()
	a.CompletedAt, a.Correct, a.Total, a.TimeTakenMs = &now, 0, len(c.QuestionIDs), &limitMs
	return true, nil
}

// ── Overview ─────────────────────────────────────────────────────────

type DailyAttemptView struct {
	Status      string            `json:"status"` // in_progress | completed
	StartedAt   time.Time         `json:"started_at"`
	Deadline    time.Time         `json:"deadline"`
	Correct     int               `json:"correct"`
	Total       int               `json:"total"`
	TimeTakenMs *int              `json:"time_taken_ms"`
	Rank        int               `json:"rank,omitempty"`
	Review      []DailyReviewItem `json:"review,omitempty"`
}

type DailyOverview struct {
	Date             string              `json:"date"`
	Available        bool                `json:"available"`
	QuestionCount    int                 `json:"question_count"`
	TimeLimitSeconds int                 `json:"time_limit_seconds"`
	ResetsAt         time.Time           `json:"resets_at"`
	Participants     int                 `json:"participants"`
	Attempt          *DailyAttemptView   `json:"attempt"`
	Leaderboard      []models.DailyEntry `json:"leaderboard"`
}

// Overview describes today's challenge for the home page: the user's own
// attempt (with a review once completed) and the top of the leaderboard.
func (s *DailyService) Overview(ctx context.Context, userID string) (*DailyOverview, error) {
	now := s.now()
	date := challengeDate(now)
	ov := &DailyOverview{
		Date:             date.Format("2006-01-02"),
		TimeLimitSeconds: DailyTimeLimitSeconds,
		ResetsAt:         nextReset(now),
		Leaderboard:      []models.DailyEntry{},
	}

	c, err := s.ensureChallenge(ctx, date)
	if errors.Is(err, ErrNoDailyChallenge) {
		return ov, nil
	}
	if err != nil {
		return nil, err
	}
	ov.Available = true
	ov.QuestionCount = len(c.QuestionIDs)
	ov.TimeLimitSeconds = c.TimeLimitSeconds

	a, err := s.dailyRepo.GetAttempt(ctx, date, userID)
	if err != nil {
		return nil, err
	}
	if a != nil {
		completed, err := s.expireIfOverdue(ctx, a, c)
		if err != nil {
			return nil, err
		}
		view := &DailyAttemptView{
			Status:      "in_progress",
			StartedAt:   a.StartedAt,
			Deadline:    deadlineOf(a, c),
			Correct:     a.Correct,
			Total:       a.Total,
			TimeTakenMs: a.TimeTakenMs,
		}
		if completed {
			view.Status = "completed"
			if a.TimeTakenMs != nil {
				view.Rank, _, _ = s.dailyRepo.Standing(ctx, date, a.Correct, *a.TimeTakenMs)
			}
			view.Review, err = s.review(ctx, c, a.Answers)
			if err != nil {
				return nil, err
			}
		}
		ov.Attempt = view
	}

	if ov.Participants, err = s.dailyRepo.CountParticipants(ctx, date); err != nil {
		return nil, err
	}
	if ov.Leaderboard, err = s.dailyRepo.Leaderboard(ctx, date, dailyLeaderboardSize); err != nil {
		return nil, err
	}
	return ov, nil
}

// ── Start ────────────────────────────────────────────────────────────

type DailyStart struct {
	Date             string                     `json:"date"`
	StartedAt        time.Time                  `json:"started_at"`
	Deadline         time.Time                  `json:"deadline"`
	ServerTime       time.Time                  `json:"server_time"` // lets clients correct for clock skew
	TimeLimitSeconds int                        `json:"time_limit_seconds"`
	Questions        []models.QuestionForPlayer `json:"questions"`
}

// Start begins (or resumes) the user's attempt and returns the questions
// without their answers. The clock starts at the first call.
func (s *DailyService) Start(ctx context.Context, userID string) (*DailyStart, error) {
	date := challengeDate(s.now())
	c, err := s.ensureChallenge(ctx, date)
	if err != nil {
		return nil, err
	}

	a, err := s.dailyRepo.StartAttempt(ctx, date, userID, len(c.QuestionIDs))
	if err != nil {
		return nil, err
	}
	completed, err := s.expireIfOverdue(ctx, a, c)
	if err != nil {
		return nil, err
	}
	if completed {
		return nil, ErrDailyAlreadyPlayed
	}

	questions, err := s.questionRepo.GetByIDsWithOptions(ctx, c.QuestionIDs)
	if err != nil {
		return nil, err
	}
	return &DailyStart{
		Date:             date.Format("2006-01-02"),
		StartedAt:        a.StartedAt,
		Deadline:         deadlineOf(a, c),
		ServerTime:       s.now(),
		TimeLimitSeconds: c.TimeLimitSeconds,
		// Keyed by attempt: resuming keeps the order, each player gets their own.
		Questions: toPlayerQuestions(questions, "daily:"+a.ID),
	}, nil
}

// ── Submit ───────────────────────────────────────────────────────────

type DailyAnswer struct {
	QuestionID string `json:"question_id"`
	OptionID   string `json:"option_id"` // empty = skipped
}

type DailyReviewItem struct {
	QuestionID         string  `json:"question_id"`
	Body               string  `json:"body"`
	SelectedOptionID   *string `json:"selected_option_id"`
	SelectedOptionText *string `json:"selected_option_text"`
	CorrectOptionID    string  `json:"correct_option_id"`
	CorrectOptionText  string  `json:"correct_option_text"`
	IsCorrect          bool    `json:"is_correct"`
	Explanation        *string `json:"explanation"`
}

type DailyResult struct {
	Correct      int               `json:"correct"`
	Total        int               `json:"total"`
	TimeTakenMs  int               `json:"time_taken_ms"`
	Expired      bool              `json:"expired"`
	Rank         int               `json:"rank"`
	Participants int               `json:"participants"`
	Review       []DailyReviewItem `json:"review"`
}

// Submit grades the user's answers. A submit after the time limit (plus a
// short grace period) scores zero.
func (s *DailyService) Submit(ctx context.Context, userID string, answers []DailyAnswer) (*DailyResult, error) {
	now := s.now()
	date := challengeDate(now)
	c, err := s.dailyRepo.GetChallenge(ctx, date)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrDailyNotStarted
	}
	a, err := s.dailyRepo.GetAttempt(ctx, date, userID)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, ErrDailyNotStarted
	}
	if a.CompletedAt != nil {
		return nil, ErrDailyAlreadyPlayed
	}

	limit := time.Duration(c.TimeLimitSeconds) * time.Second
	elapsed := now.Sub(a.StartedAt)
	expired := elapsed > limit+dailyGrace
	if elapsed > limit {
		elapsed = limit
	}

	questions, err := s.questionRepo.GetByIDsWithOptions(ctx, c.QuestionIDs)
	if err != nil {
		return nil, err
	}
	chosen := make(map[string]string, len(answers))
	for _, ans := range answers {
		if ans.OptionID != "" {
			chosen[ans.QuestionID] = ans.OptionID
		}
	}

	records := make([]models.DailyAnswerRecord, 0, len(questions))
	correct := 0
	for _, q := range questions {
		rec := models.DailyAnswerRecord{QuestionID: q.ID}
		if opt, ok := chosen[q.ID]; ok && !expired {
			o := opt
			rec.OptionID = &o
			rec.IsCorrect = correctOptionID(q) == opt
		}
		if rec.IsCorrect {
			correct++
		}
		records = append(records, rec)
	}

	timeMs := int(elapsed.Milliseconds())
	ok, err := s.dailyRepo.CompleteAttempt(ctx, a.ID, correct, len(questions), timeMs, records)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrDailyAlreadyPlayed
	}

	rank, participants, err := s.dailyRepo.Standing(ctx, date, correct, timeMs)
	if err != nil {
		return nil, err
	}
	review, err := s.reviewFrom(questions, records)
	if err != nil {
		return nil, err
	}
	return &DailyResult{
		Correct:      correct,
		Total:        len(questions),
		TimeTakenMs:  timeMs,
		Expired:      expired,
		Rank:         rank,
		Participants: participants,
		Review:       review,
	}, nil
}

func correctOptionID(q models.Question) string {
	for _, o := range q.Options {
		if o.IsCorrect {
			return o.ID
		}
	}
	return ""
}

func optionText(q models.Question, optionID string) string {
	for _, o := range q.Options {
		if o.ID == optionID {
			return o.OptionText
		}
	}
	return ""
}

func (s *DailyService) review(ctx context.Context, c *models.DailyChallenge, records []models.DailyAnswerRecord) ([]DailyReviewItem, error) {
	questions, err := s.questionRepo.GetByIDsWithOptions(ctx, c.QuestionIDs)
	if err != nil {
		return nil, fmt.Errorf("load daily review: %w", err)
	}
	return s.reviewFrom(questions, records)
}

func (s *DailyService) reviewFrom(questions []models.Question, records []models.DailyAnswerRecord) ([]DailyReviewItem, error) {
	byQuestion := make(map[string]models.DailyAnswerRecord, len(records))
	for _, r := range records {
		byQuestion[r.QuestionID] = r
	}
	items := make([]DailyReviewItem, 0, len(questions))
	for _, q := range questions {
		rec := byQuestion[q.ID]
		correctID := correctOptionID(q)
		item := DailyReviewItem{
			QuestionID:        q.ID,
			Body:              q.Body,
			SelectedOptionID:  rec.OptionID,
			CorrectOptionID:   correctID,
			CorrectOptionText: optionText(q, correctID),
			IsCorrect:         rec.IsCorrect,
			Explanation:       q.Explanation,
		}
		if rec.OptionID != nil {
			text := optionText(q, *rec.OptionID)
			item.SelectedOptionText = &text
		}
		items = append(items, item)
	}
	return items, nil
}
