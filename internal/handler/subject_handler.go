package handler

import (
	"net/http"

	"github.com/exam-arena/internal/repository"
	"github.com/exam-arena/internal/utils"
)

type SubjectHandler struct {
	questionRepo *repository.QuestionRepo
}

func NewSubjectHandler(questionRepo *repository.QuestionRepo) *SubjectHandler {
	return &SubjectHandler{questionRepo: questionRepo}
}

func (h *SubjectHandler) GetAllSubjects(w http.ResponseWriter, r *http.Request) {
	subjects, err := h.questionRepo.GetActiveCategories(r.Context())
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to load subjects")
		return
	}

	utils.JSON(w, http.StatusOK, subjects)
}
