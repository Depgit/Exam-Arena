package questions

import (
	"net/http"

	"github.com/exam-arena/internal/platform/respond"
)

type SubjectHandler struct {
	questionRepo *Store
}

func NewSubjectHandler(questionRepo *Store) *SubjectHandler {
	return &SubjectHandler{questionRepo: questionRepo}
}

func (h *SubjectHandler) GetAllSubjects(w http.ResponseWriter, r *http.Request) {
	subjects, err := h.questionRepo.GetActiveCategories(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to load subjects")
		return
	}

	respond.JSON(w, http.StatusOK, subjects)
}
