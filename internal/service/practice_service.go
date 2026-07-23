package service

import (
	"context"
	"fmt"

	"github.com/exam-arena/internal/models"
	"github.com/exam-arena/internal/repository"
)

type PracticeService struct {
	practiceRepo *repository.PracticeRepo
	questionRepo *repository.QuestionRepo
}

func NewPracticeService(practiceRepo *repository.PracticeRepo, questionRepo *repository.QuestionRepo) *PracticeService {
	return &PracticeService{
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

func (s *PracticeService) StartSession(ctx context.Context, userID string, req StartPracticeRequest) (*PracticeSessionResponse, error) {
	if req.QuestionCount <= 0 || req.QuestionCount > 50 {
		req.QuestionCount = 10
	}

	// Create session
	session, err := s.practiceRepo.CreateSession(ctx, userID, req.ExamCategoryID, req.TopicID, req.Difficulty, req.QuestionCount)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	// Get random questions
	questions, err := s.questionRepo.GetRandomQuestions(ctx, req.ExamCategoryID, req.QuestionCount)
	if err != nil {
		return nil, fmt.Errorf("get questions: %w", err)
	}

	// Add questions to session
	playerQuestions := make([]models.QuestionForPlayer, len(questions))
	for i, q := range questions {
		if err := s.practiceRepo.AddSessionQuestion(ctx, session.ID, q.ID, i+1); err != nil {
			return nil, err
		}

		opts := make([]models.OptionForPlayer, len(q.Options))
		for j, o := range q.Options {
			opts[j] = models.OptionForPlayer{
				ID:         o.ID,
				OptionText: o.OptionText,
				OrderIndex: o.OrderIndex,
			}
		}
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

func (s *PracticeService) SubmitAnswer(ctx context.Context, userID, sessionID string, req SubmitPracticeAnswerRequest) (*PracticeAnswerResponse, error) {
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

func (s *PracticeService) EndSession(ctx context.Context, userID, sessionID string) error {
	session, err := s.practiceRepo.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	if session == nil || session.UserID != userID {
		return fmt.Errorf("session not found")
	}
	return s.practiceRepo.EndSession(ctx, sessionID)
}

func (s *PracticeService) GetSession(ctx context.Context, userID, sessionID string) (*repository.PracticeSession, error) {
	session, err := s.practiceRepo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil || session.UserID != userID {
		return nil, fmt.Errorf("session not found")
	}
	return session, nil
}
