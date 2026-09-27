package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/exam-arena/internal/models"
	"github.com/exam-arena/internal/repository"
	"github.com/exam-arena/internal/utils"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func isUUID(s string) bool { return uuidPattern.MatchString(s) }

const maxTopicNameLen = 100 // topics.name is VARCHAR(100)

type TopicHandler struct {
	topicRepo *repository.TopicRepo
}

func NewTopicHandler(topicRepo *repository.TopicRepo) *TopicHandler {
	return &TopicHandler{topicRepo: topicRepo}
}

type CreateTopicRequest struct {
	ExamCategoryID string  `json:"exam_category_id"`
	Name           string  `json:"name"`
	ParentTopicID  *string `json:"parent_topic_id"`
}

// ListTopics godoc
// GET /api/v1/topics?exam_category_id=<uuid>
// GET /api/v1/subjects/{id}/topics
// Returns every topic, or only one category's topics when a category is
// given (path param on the subject-scoped route, query param otherwise).
// Admins use it to pick a topic_id when creating a question; players can
// use it to filter practice sessions.
func (h *TopicHandler) ListTopics(w http.ResponseWriter, r *http.Request) {
	categoryID := r.PathValue("id")
	if categoryID == "" {
		categoryID = r.URL.Query().Get("exam_category_id")
	}
	if categoryID != "" && !isUUID(categoryID) {
		utils.JSONError(w, http.StatusBadRequest, "exam_category_id must be a UUID")
		return
	}

	topics, err := h.topicRepo.List(r.Context(), categoryID)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to load topics")
		return
	}

	utils.JSONWithMeta(w, http.StatusOK, topics, map[string]interface{}{
		"count":            len(topics),
		"exam_category_id": categoryID,
	})
}

// GetTopic godoc
// GET /api/v1/topics/{id}
func (h *TopicHandler) GetTopic(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !isUUID(id) {
		utils.JSONError(w, http.StatusBadRequest, "topic id must be a UUID")
		return
	}

	topic, err := h.topicRepo.GetByID(r.Context(), id)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to load topic")
		return
	}
	if topic == nil {
		utils.JSONError(w, http.StatusNotFound, "topic not found")
		return
	}

	utils.JSON(w, http.StatusOK, topic)
}

// CreateTopic godoc
// POST /api/v1/admin/topics   (admin only)
// Body: { "exam_category_id": "<uuid>", "name": "Number Series", "parent_topic_id": "<uuid>|null" }
func (h *TopicHandler) CreateTopic(w http.ResponseWriter, r *http.Request) {
	var req CreateTopicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	switch {
	case req.ExamCategoryID == "":
		utils.JSONError(w, http.StatusBadRequest, "exam_category_id is required")
		return
	case !isUUID(req.ExamCategoryID):
		utils.JSONError(w, http.StatusBadRequest, "exam_category_id must be a UUID")
		return
	case req.Name == "":
		utils.JSONError(w, http.StatusBadRequest, "name is required")
		return
	case len(req.Name) > maxTopicNameLen:
		utils.JSONError(w, http.StatusBadRequest, "name must be at most 100 characters")
		return
	}

	if req.ParentTopicID != nil {
		if *req.ParentTopicID == "" {
			req.ParentTopicID = nil
		} else {
			if !isUUID(*req.ParentTopicID) {
				utils.JSONError(w, http.StatusBadRequest, "parent_topic_id must be a UUID")
				return
			}
			parent, err := h.topicRepo.GetByID(r.Context(), *req.ParentTopicID)
			if err != nil {
				utils.JSONError(w, http.StatusInternalServerError, "failed to load parent topic")
				return
			}
			if parent == nil {
				utils.JSONError(w, http.StatusBadRequest, "parent topic not found")
				return
			}
			if parent.ExamCategoryID != req.ExamCategoryID {
				utils.JSONError(w, http.StatusBadRequest, "parent topic belongs to a different exam category")
				return
			}
		}
	}

	exists, err := h.topicRepo.ExistsByName(r.Context(), req.ExamCategoryID, req.Name)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to check topic name")
		return
	}
	if exists {
		utils.JSONError(w, http.StatusConflict, "a topic with this name already exists in this category")
		return
	}

	topic, err := h.topicRepo.Create(r.Context(), &models.Topic{
		ExamCategoryID: req.ExamCategoryID,
		ParentTopicID:  req.ParentTopicID,
		Name:           req.Name,
	})
	if err != nil {
		if repository.IsForeignKeyViolation(err) {
			utils.JSONError(w, http.StatusBadRequest, "exam category not found")
			return
		}
		utils.JSONError(w, http.StatusInternalServerError, "failed to create topic")
		return
	}

	utils.JSON(w, http.StatusCreated, topic)
}
