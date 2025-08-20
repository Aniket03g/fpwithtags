package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

// Test data structures
type CreateReleaseRequest struct {
	Tag   string `json:"tag"`
	PRs   []int  `json:"prs"`
	Notes string `json:"notes"`
}

func main() {
	baseURL := "http://localhost:8080"
	
	// Test 1: Approve a PR
	prID := 1 // Change this to an existing PR ID in your database
	fmt.Printf("Testing ApprovePR endpoint for PR ID %d...\n", prID)
	
	approveResp, err := http.Post(
		fmt.Sprintf("%s/prs/%d/approve", baseURL, prID),
		"application/json",
		nil,
	)
	
	if err != nil {
		fmt.Printf("Error calling ApprovePR endpoint: %v\n", err)
		os.Exit(1)
	}
	defer approveResp.Body.Close()
	
	fmt.Printf("ApprovePR response status: %s\n", approveResp.Status)
	
	// Test 2: Create a Release
	fmt.Println("\nTesting CreateRelease endpoint...")
	
	releaseReq := CreateReleaseRequest{
		Tag:   "v1.0.0",
		PRs:   []int{prID}, // Using the same PR we just approved
		Notes: "Test release created via test script",
	}
	
	reqBody, err := json.Marshal(releaseReq)
	if err != nil {
		fmt.Printf("Error marshaling release request: %v\n", err)
		os.Exit(1)
	}
	
	releaseResp, err := http.Post(
		fmt.Sprintf("%s/api/releases", baseURL),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	
	if err != nil {
		fmt.Printf("Error calling CreateRelease endpoint: %v\n", err)
		os.Exit(1)
	}
	defer releaseResp.Body.Close()
	
	fmt.Printf("CreateRelease response status: %s\n", releaseResp.Status)
	
	// Parse the response to get the release ID
	var releaseResult map[string]interface{}
	if err := json.NewDecoder(releaseResp.Body).Decode(&releaseResult); err != nil {
		fmt.Printf("Error parsing release response: %v\n", err)
	} else {
		if releaseID, ok := releaseResult["release_id"].(float64); ok {
			fmt.Printf("Created release with ID: %d\n", int(releaseID))
		}
	}
	
	fmt.Println("\nTest completed!")
}
