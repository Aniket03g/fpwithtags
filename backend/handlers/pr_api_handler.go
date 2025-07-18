package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PRUploadRequest struct {
	FeatureID   int    `json:"feature_id"`
	TaskID      int    `json:"task_id"`
	PRURL       string `json:"pr_url"`
	Branch      string `json:"branch"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IsTested    bool   `json:"is_tested"`
}

func PRUploadAPIHandler(c *gin.Context) {
	var req PRUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Print the received JSON to the backend terminal
	fmt.Printf("Received PR JSON from CLI: %+v\n", req)
	c.JSON(http.StatusOK, gin.H{"status": "received", "data": req})
}
