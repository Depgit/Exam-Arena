package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/exam-arena/internal/service"
	"github.com/exam-arena/internal/utils"
)

// GeneratorHandler exposes the question generator to admins: a preview to
// check question quality, and a manual pool rotation.
type GeneratorHandler struct {
	pool *service.QuestionPool
}

func NewGeneratorHandler(pool *service.QuestionPool) *GeneratorHandler {
	return &GeneratorHandler{pool: pool}
}

// Preview generates sample questions with their answers. Nothing is stored.
//
//	GET /api/v1/admin/generator/preview?category=MATH&difficulty=hard&n=20
func (h *GeneratorHandler) Preview(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	category := strings.ToUpper(q.Get("category"))
	difficulty := strings.ToLower(q.Get("difficulty"))
	switch difficulty {
	case "", "easy", "medium", "hard":
	default:
		utils.JSONError(w, http.StatusBadRequest, "difficulty must be easy, medium or hard")
		return
	}
	n, _ := strconv.Atoi(q.Get("n"))

	questions, err := h.pool.Preview(category, difficulty, n)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, questions)
}

// Rotate retires part of each generated pool and refills it now, instead
// of waiting for the next scheduled rotation.
//
//	POST /api/v1/admin/generator/rotate
func (h *GeneratorHandler) Rotate(w http.ResponseWriter, r *http.Request) {
	res, err := h.pool.RotateNow(r.Context())
	if err != nil {
		slog.Error("manual question pool rotation failed", "error", err)
		utils.JSONError(w, http.StatusInternalServerError, "rotation failed")
		return
	}
	utils.JSON(w, http.StatusOK, res)
}
