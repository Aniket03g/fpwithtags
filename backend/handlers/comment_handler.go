package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
	commentRepo    *repositories.CommentRepository
	attachmentRepo repositories.TaskAttachmentRepository
}

func NewCommentHandler(commentRepo *repositories.CommentRepository, attachmentRepo repositories.TaskAttachmentRepository) *CommentHandler {
	return &CommentHandler{
		commentRepo:    commentRepo,
		attachmentRepo: attachmentRepo,
	}
}

// CreateComment creates a new comment for a task
func (h *CommentHandler) CreateComment(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("task_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var comment models.Comment
	if err := c.ShouldBind(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment.TaskID = uint(taskID)

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	comment.UserID = userID.(uint)

	if err := h.commentRepo.Create(&comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	// Check if this is an HTMX request
	if c.GetHeader("HX-Request") == "true" {
		// Return HTML for HTMX
		tmpl, err := template.ParseFiles("templates/_comment-item.html")
		if err != nil {
			log.Printf("template parse error: %v", err)
			c.String(http.StatusInternalServerError, "Template error")
			return
		}

		// Add author name for display
		commentData := map[string]interface{}{
			"ID":         comment.ID,
			"Content":    comment.Content,
			"AuthorName": fmt.Sprintf("User #%d", comment.UserID),
			"CreatedAt":  comment.CreatedAt,
			"UpdatedAt":  comment.UpdatedAt,
		}

		err = tmpl.ExecuteTemplate(c.Writer, "comment-item", commentData)
		if err != nil {
			log.Printf("template execute error: %v", err)
			c.String(http.StatusInternalServerError, "Render error")
		}
	} else {
		// Return JSON for API
		c.JSON(http.StatusCreated, comment)
	}
}

// GetTaskComments retrieves all comments for a task
func (h *CommentHandler) GetTaskComments(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("task_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	comments, err := h.commentRepo.GetByTaskID(uint(taskID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch comments"})
		return
	}

	c.JSON(http.StatusOK, comments)
}

// GetAttachmentComments retrieves all comments for an attachment
func (h *CommentHandler) GetAttachmentComments(c *gin.Context) {
	attachmentID, err := strconv.ParseUint(c.Param("attachment_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attachment ID"})
		return
	}

	comments, err := h.commentRepo.GetByAttachmentID(uint(attachmentID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch comments"})
		return
	}

	c.JSON(http.StatusOK, comments)
}

// UpdateComment updates an existing comment
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("comment_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	var comment models.Comment
	if err := c.ShouldBind(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment.ID = uint(commentID)

	if err := h.commentRepo.Update(&comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment"})
		return
	}

	// Check if this is an HTMX request
	if c.GetHeader("HX-Request") == "true" {
		// Return HTML for HTMX
		tmpl, err := template.ParseFiles("templates/_comment-item.html")
		if err != nil {
			log.Printf("template parse error: %v", err)
			c.String(http.StatusInternalServerError, "Template error")
			return
		}

		// Add author name for display
		commentData := map[string]interface{}{
			"ID":         comment.ID,
			"Content":    comment.Content,
			"AuthorName": fmt.Sprintf("User #%d", comment.UserID),
			"CreatedAt":  comment.CreatedAt,
			"UpdatedAt":  comment.UpdatedAt,
		}

		err = tmpl.ExecuteTemplate(c.Writer, "comment-item", commentData)
		if err != nil {
			log.Printf("template execute error: %v", err)
			c.String(http.StatusInternalServerError, "Render error")
		}
	} else {
		// Return JSON for API
		c.JSON(http.StatusOK, comment)
	}
}

// DeleteComment deletes a comment
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("comment_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	if err := h.commentRepo.Delete(uint(commentID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}

	// For HTMX requests, return empty body to remove the element
	if c.GetHeader("HX-Request") == "true" {
		c.Status(http.StatusOK)
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Comment deleted"})
	}
}
