package practice

import (
	"context"
	"fmt"

	"github.com/exam-arena/internal/models"
	"github.com/exam-arena/internal/questions"
)

type Service struct {
	practiceRepo *Store
	questionRepo *questions.Store
}

func NewService(practiceRepo *Store, questionRepo *questions.Store) *Service {
	return &Service{
		practiceRepo: practiceRepo,
		questionRepo: questionRepo,
	}
}

type StartPracticeRequest struct {
	ExamCategoryID string  `json:"exam_category_id"`
	TopicID        *string `json:"topic_id"`
	Difficulty     *string `json:"difficulty"`
	QuestionCount  int     `json:"question_count"`
}

type PracticeSessionResponse struct {
	SessionID string                     `json:"session_id"`
	Questions []models.QuestionForPlayer `json:"questions"`
}

func (s *Service) StartSession(ctx context.Context, userID string, req StartPracticeRequest) (*PracticeSessionResponse, error) {
	if req.QuestionCount <= 0 || req.QuestionCount > 50 {
		req.QuestionCount = 10
	}

	if req.Difficulty != nil {
		switch *req.Difficulty {
		case "", "easy", "medium", "hard":
		default:
			return nil, fmt.Errorf("invalid difficulty %q", *req.Difficulty)
		}
	}

	// Create session
	session, err := s.practiceRepo.CreateSession(ctx, userID, req.ExamCategoryID, req.TopicID, req.Difficulty, req.QuestionCount)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	// Get random questions
	picked, err := s.questionRepo.GetRandomQuestions(ctx, req.ExamCategoryID, req.Difficulty, req.QuestionCount)
	if err != nil {
		return nil, fmt.Errorf("get questions: %w", err)
	}

	// Add questions to session
	playerQuestions := make([]models.QuestionForPlayer, len(picked))
	for i, q := range picked {
		if err := s.practiceRepo.AddSessionQuestion(ctx, session.ID, q.ID, i+1); err != nil {
			return nil, err
		}

		opts := questions.PlayerOptions(q, "practice:"+session.ID)
		playerQuestions[i] = models.QuestionForPlayer{
			ID:                   q.ID,
			QuestionType:         q.QuestionType,
			Difficulty:           q.Difficulty,
			Body:                 q.Body,
			EstimatedTimeSeconds: q.EstimatedTimeSeconds,
			Options:              opts,
			OrderIndex:           i + 1,
		}
	}

	return &PracticeSessionResponse{
		SessionID: session.ID,
		Questions: playerQuestions,
	}, nil
}

type SubmitPracticeAnswerRequest struct {
	QuestionID  string `json:"question_id"`
	OptionID    string `json:"option_id"`
	TimeTakenMs int    `json:"time_taken_ms"`
}

type PracticeAnswerResponse struct {
	IsCorrect   bool    `json:"is_correct"`
	Explanation *string `json:"explanation"`
}

func (s *Service) SubmitAnswer(ctx context.Context, userID, sessionID string, req SubmitPracticeAnswerRequest) (*PracticeAnswerResponse, error) {
	session, err := s.practiceRepo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil || session.UserID != userID {
		return nil, fmt.Errorf("session not found")
	}

	isCorrect, err := s.questionRepo.IsOptionCorrect(ctx, req.OptionID)
	if err != nil {
		return nil, err
	}

	if err := s.practiceRepo.AnswerQuestion(ctx, sessionID, req.QuestionID, &req.OptionID, isCorrect, req.TimeTakenMs); err != nil {
		return nil, err
	}

	question, _ := s.questionRepo.GetByID(ctx, req.QuestionID)
	var explanation *string
	if question != nil {
		explanation = question.Explanation
	}

	return &PracticeAnswerResponse{
		IsCorrect:   isCorrect,
		Explanation: explanation,
	}, nil
}

func (s *Service) EndSession(ctx context.Context, userID, sessionID string) error {
	session, err := s.practiceRepo.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	if session == nil || session.UserID != userID {
		return fmt.Errorf("session not found")
	}
	return s.practiceRepo.EndSession(ctx, sessionID)
}

func (s *Service) GetSession(ctx context.Context, userID, sessionID string) (*Session, error) {
	session, err := s.practiceRepo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil || session.UserID != userID {
		return nil, fmt.Errorf("session not found")
	}
	return session, nil
}
