package handler

import (
	"encoding/json"
	"net/http"

	"github.com/exam-arena/internal/models"
	"github.com/exam-arena/internal/repository"
	"github.com/exam-arena/internal/utils"
)

type AdminHandler struct {
	questionRepo *repository.QuestionRepo
	userRepo     *repository.UserRepo
	matchRepo    *repository.MatchRepo
}

func NewAdminHandler(questionRepo *repository.QuestionRepo, userRepo *repository.UserRepo, matchRepo *repository.MatchRepo) *AdminHandler {
	return &AdminHandler{
		questionRepo: questionRepo,
		userRepo:     userRepo,
		matchRepo:    matchRepo,
	}
}

type CreateQuestionRequest struct {
	ExamCategoryID       string                `json:"exam_category_id"`
	TopicID              string                `json:"topic_id"`
	QuestionType         string                `json:"question_type"`
	Difficulty           string                `json:"difficulty"`
	Body                 string                `json:"body"`
	Explanation          *string               `json:"explanation"`
	EstimatedTimeSeconds int                   `json:"estimated_time_seconds"`
	Options              []CreateOptionRequest `json:"options"`
}

type CreateOptionRequest struct {
	OptionText string `json:"option_text"`
	IsCorrect  bool   `json:"is_correct"`
}

func (h *AdminHandler) CreateQuestion(w http.ResponseWriter, r *http.Request) {
	var req CreateQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Body == "" || req.ExamCategoryID == "" || req.TopicID == "" {
		utils.JSONError(w, http.StatusBadRequest, "missing required fields")
		return
	}
	if len(req.Options) < 2 {
		utils.JSONError(w, http.StatusBadRequest, "at least 2 options required")
		return
	}

	if req.EstimatedTimeSeconds == 0 {
		req.EstimatedTimeSeconds = 60
	}
	if req.QuestionType == "" {
		req.QuestionType = "mcq_single"
	}

	q := &models.Question{
		ExamCategoryID:       req.ExamCategoryID,
		TopicID:              req.TopicID,
		QuestionType:         req.QuestionType,
		Difficulty:           req.Difficulty,
		Body:                 req.Body,
		Explanation:          req.Explanation,
		EstimatedTimeSeconds: req.EstimatedTimeSeconds,
		Status:               "draft",
	}

	options := make([]models.QuestionOption, len(req.Options))
	for i, o := range req.Options {
		options[i] = models.QuestionOption{
			OptionText: o.OptionText,
			IsCorrect:  o.IsCorrect,
		}
	}

	created, err := h.questionRepo.Create(r.Context(), q, options)
	if err != nil {
		switch {
		case repository.IsForeignKeyViolation(err):
			utils.JSONError(w, http.StatusBadRequest, "exam_category_id or topic_id does not exist")
		case repository.IsInvalidInput(err):
			utils.JSONError(w, http.StatusBadRequest, "invalid field value (check UUIDs, question_type and difficulty)")
		default:
			utils.JSONError(w, http.StatusInternalServerError, "failed to create question")
		}
		return
	}

	utils.JSON(w, http.StatusCreated, created)
}

func (h *AdminHandler) PublishQuestion(w http.ResponseWriter, r *http.Request) {
	questionID := r.PathValue("id")
	if err := h.questionRepo.Publish(r.Context(), questionID); err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to publish question")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "published"})
}

func (h *AdminHandler) GetSystemStats(w http.ResponseWriter, r *http.Request) {
	userCount, _ := h.userRepo.GetUserCount(r.Context())
	questionCount, _ := h.questionRepo.GetQuestionCount(r.Context())
	activeMatches, _ := h.matchRepo.GetActiveMatchCount(r.Context())
	totalMatches, _ := h.matchRepo.GetTotalMatchCount(r.Context())

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"users":          userCount,
		"questions":      questionCount,
		"active_matches": activeMatches,
		"total_matches":  totalMatches,
	})
}
