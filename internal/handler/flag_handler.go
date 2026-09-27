package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	appCache "github.com/exam-arena/internal/cache"
	"github.com/exam-arena/internal/middleware"
	"github.com/exam-arena/internal/repository"
	"github.com/exam-arena/internal/utils"
)

// FlagReasons are the reasons a player can give when flagging a question.
var FlagReasons = map[string]bool{
	"wrong_answer":     true, // the marked correct answer is wrong
	"multiple_correct": true, // more than one option is correct
	"unclear":          true, // question or options are ambiguous
	"typo":             true, // spelling / formatting mistake
	"other":            true, // anything else (description required)
}

const maxFlagDescription = 500

type FlagHandler struct {
	flagRepo     *repository.FlagRepo
	questionRepo *repository.QuestionRepo
	questionBank *appCache.QuestionBank
}

func NewFlagHandler(flagRepo *repository.FlagRepo, questionRepo *repository.QuestionRepo, questionBank *appCache.QuestionBank) *FlagHandler {
	return &FlagHandler{flagRepo: flagRepo, questionRepo: questionRepo, questionBank: questionBank}
}

// FlagQuestion godoc
// POST /api/v1/questions/{id}/flag
// Body: { "reason": "wrong_answer", "description": "Option B is also correct" }
func (h *FlagHandler) FlagQuestion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Reason      string `json:"reason"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Description = strings.TrimSpace(req.Description)
	if !FlagReasons[req.Reason] {
		utils.JSONError(w, http.StatusBadRequest, "reason must be one of wrong_answer, multiple_correct, unclear, typo, other")
		return
	}
	if req.Reason == "other" && req.Description == "" {
		utils.JSONError(w, http.StatusBadRequest, "please describe the problem")
		return
	}
	if len(req.Description) > maxFlagDescription {
		utils.JSONError(w, http.StatusBadRequest, "description must be at most 500 characters")
		return
	}

	questionID := r.PathValue("id")
	q, err := h.questionRepo.GetByID(r.Context(), questionID)
	if err != nil && !repository.IsInvalidInput(err) {
		slog.Error("flag: load question", "error", err)
		utils.JSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if q == nil {
		utils.JSONError(w, http.StatusNotFound, "question not found")
		return
	}

	var description *string
	if req.Description != "" {
		description = &req.Description
	}
	flag, err := h.flagRepo.CreateQuestionFlag(r.Context(), middleware.GetUserID(r), q.ID, req.Reason, description)
	if err != nil {
		if repository.IsUniqueViolation(err) {
			utils.JSONError(w, http.StatusConflict, "you have already flagged this question")
			return
		}
		slog.Error("flag: create", "error", err)
		utils.JSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	utils.JSON(w, http.StatusCreated, flag)
}

// ListFlags godoc
// GET /api/v1/admin/flags?status=open|reviewed|resolved|dismissed|all  (default open)
func (h *FlagHandler) ListFlags(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	switch status {
	case "":
		status = "open"
	case "open", "reviewed", "resolved", "dismissed", "all":
	default:
		utils.JSONError(w, http.StatusBadRequest, "status must be open, reviewed, resolved, dismissed or all")
		return
	}

	flagged, err := h.flagRepo.ListFlaggedQuestions(r.Context(), status)
	if err != nil {
		slog.Error("list flags", "error", err)
		utils.JSONError(w, http.StatusInternalServerError, "failed to load flags")
		return
	}

	// Attach options (with the correct answer) so admins can judge the flag.
	ids := make([]string, len(flagged))
	for i, f := range flagged {
		ids[i] = f.QuestionID
	}
	questions, err := h.questionRepo.GetByIDsWithOptions(r.Context(), ids)
	if err == nil {
		byID := make(map[string]int, len(flagged))
		for i, f := range flagged {
			byID[f.QuestionID] = i
		}
		for _, q := range questions {
			if q.Options != nil {
				flagged[byID[q.ID]].Options = q.Options
			}
		}
	}

	utils.JSONWithMeta(w, http.StatusOK, flagged, map[string]interface{}{
		"status": status,
		"count":  len(flagged),
	})
}

// ReviewFlags godoc
// PUT /api/v1/admin/flags/{questionId}
// Body: { "status": "resolved" | "dismissed" | "reviewed", "archive_question": false }
//
// Closes every open flag on the question. archive_question also takes the
// question out of play immediately.
func (h *FlagHandler) ReviewFlags(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Status          string `json:"status"`
		ArchiveQuestion bool   `json:"archive_question"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	switch req.Status {
	case "resolved", "dismissed", "reviewed":
	default:
		utils.JSONError(w, http.StatusBadRequest, "status must be resolved, dismissed or reviewed")
		return
	}

	q, err := h.questionRepo.GetByID(r.Context(), r.PathValue("questionId"))
	if err != nil && !repository.IsInvalidInput(err) {
		utils.JSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if q == nil {
		utils.JSONError(w, http.StatusNotFound, "question not found")
		return
	}

	closed, err := h.flagRepo.ReviewQuestionFlags(r.Context(), q.ID, middleware.GetUserID(r), req.Status)
	if err != nil {
		slog.Error("review flags", "error", err)
		utils.JSONError(w, http.StatusInternalServerError, "failed to update flags")
		return
	}

	if req.ArchiveQuestion {
		if err := h.questionRepo.Archive(r.Context(), q.ID); err != nil {
			utils.JSONError(w, http.StatusInternalServerError, "failed to archive question")
			return
		}
		if err := h.questionBank.RefreshCategory(r.Context(), q.ExamCategoryID); err != nil {
			slog.Warn("archive: refresh question bank", "error", err)
		}
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"question_id":       q.ID,
		"flags_closed":      closed,
		"status":            req.Status,
		"question_archived": req.ArchiveQuestion,
	})
}
