package handlers

import (
	"net/http"
	"strconv"

	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/gin-gonic/gin"
)

type PullRequestHandler struct {
	prRepo repositories.PullRequestRepository
}

func NewPullRequestHandler(prRepo repositories.PullRequestRepository) *PullRequestHandler {
	return &PullRequestHandler{prRepo: prRepo}
}

// POST /tasks/:id/prs
func (h *PullRequestHandler) AddPullRequest(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid task ID")
		return
	}
	var pr models.PullRequest
	if err := c.ShouldBind(&pr); err != nil {
		c.String(http.StatusBadRequest, "Invalid form data")
		return
	}
	pr.TaskID = uint(taskID)
	pr.Status = "Open"
	pr.Tested = false
	if err := h.prRepo.Create(&pr); err != nil {
		c.String(http.StatusInternalServerError, "Failed to save PR")
		return
	}
	// Return empty modal (close it)
	c.HTML(http.StatusOK, "", nil)
}

// POST /prs/:id/mark-tested
func (h *PullRequestHandler) MarkTested(c *gin.Context) {
	prID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid PR ID")
		return
	}
	if err := h.prRepo.MarkTested(uint(prID)); err != nil {
		c.String(http.StatusInternalServerError, "Failed to mark as tested")
		return
	}
	c.String(http.StatusOK, "OK")
}
