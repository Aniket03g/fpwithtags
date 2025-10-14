package services

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/FeaturePlus/backend/internal/llm"
)

// FeatureSuggestion represents a suggested feature from the LLM
type FeatureSuggestion struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Priority    string `json:"priority"`
}

// DEPRECATED: Hugging Face integration (replaced with Gemini)
// HuggingFaceRequest represents the request payload for Hugging Face API
// type HuggingFaceRequest struct {
// 	Inputs     string                 `json:"inputs"`
// 	Parameters map[string]interface{} `json:"parameters,omitempty"`
// }

// HuggingFaceResponse represents the response from Hugging Face API
// type HuggingFaceResponse []struct {
// 	GeneratedText string `json:"generated_text"`
// }

// LLMService handles LLM-based feature generation
type LLMService struct {
	// Removed Hugging Face specific fields
}

// NewLLMService creates a new LLM service instance
func NewLLMService() *LLMService {
	// Refactored to use Gemini instead of Hugging Face
	log.Printf("INFO: Using Google Gemini API for LLM suggestions")
	return &LLMService{}
}

// GenerateFeatureSuggestions generates feature suggestions using Google Gemini
func (s *LLMService) GenerateFeatureSuggestions(projectContext string, techStack string) ([]FeatureSuggestion, error) {
	// Build the prompt
	prompt := s.buildPrompt(projectContext, techStack)
	log.Printf("INFO: LLM prompt length: %d characters", len(prompt))

	// Call Gemini API (replaces Hugging Face)
	response, err := llm.GenerateWithGemini(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call Gemini API: %w", err)
	}

	// Parse the response
	suggestions, err := s.parseResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	log.Printf("INFO: Generated %d feature suggestions", len(suggestions))
	return suggestions, nil
}

// buildPrompt constructs the prompt for the LLM
func (s *LLMService) buildPrompt(projectContext string, techStack string) string {
	if projectContext == "" {
		projectContext = "A software project"
	}
	if techStack == "" {
		techStack = "General"
	}

	// Stage 4b: Optimized prompt for Mistral model
	prompt := fmt.Sprintf(`You are an assistant that suggests next actionable features for the following project:

Project Context: %s
Tech Stack: %s

List 5 clear feature suggestions in JSON format. Each feature must have:
- name: Feature name
- description: Brief description
- category: One of (Backend, Frontend, Database, Auth, API, UI, Testing)
- priority: One of (High, Medium, Low)

Respond with ONLY a valid JSON array:
[{"name":"Feature 1","description":"...","category":"Backend","priority":"High"}]

JSON:`, projectContext, techStack)

	return prompt
}

// DEPRECATED: Hugging Face API call (replaced with Gemini)
// callHuggingFaceAPI makes the HTTP request to Hugging Face
// func (s *LLMService) callHuggingFaceAPI(prompt string) (string, error) {
// 	// Prepare request payload
// 	requestBody := HuggingFaceRequest{
// 		Inputs: prompt,
// 		Parameters: map[string]interface{}{
// 			"max_new_tokens": 500,
// 			"temperature":    0.7,
// 			"top_p":          0.9,
// 			"return_full_text": false,
// 		},
// 	}
//
// 	jsonData, err := json.Marshal(requestBody)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to marshal request: %w", err)
// 	}
//
// 	// Create HTTP request
// 	req, err := http.NewRequest("POST", s.apiURL, bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		return "", fmt.Errorf("failed to create request: %w", err)
// 	}
//
// 	req.Header.Set("Authorization", "Bearer "+s.apiKey)
// 	req.Header.Set("Content-Type", "application/json")
//
// 	// Send request
// 	client := &http.Client{Timeout: 30 * time.Second}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to send request: %w", err)
// 	}
// 	defer resp.Body.Close()
//
// 	// Read response
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to read response: %w", err)
// 	}
//
// 	if resp.StatusCode != http.StatusOK {
// 		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
// 	}
//
// 	// Log raw response for debugging (Stage 4b)
// 	log.Printf("DEBUG: Hugging Face raw response: %s", string(body))
//
// 	// Parse Hugging Face response format
// 	var hfResponse HuggingFaceResponse
// 	if err := json.Unmarshal(body, &hfResponse); err != nil {
// 		// If it's not in the expected format, return the raw body
// 		log.Printf("DEBUG: Response is not in expected format, using raw body")
// 		return string(body), nil
// 	}
//
// 	if len(hfResponse) > 0 {
// 		log.Printf("DEBUG: Extracted generated text from response")
// 		return hfResponse[0].GeneratedText, nil
// 	}
//
// 	return string(body), nil
// }

// parseResponse attempts to parse the LLM response into feature suggestions
func (s *LLMService) parseResponse(response string) ([]FeatureSuggestion, error) {
	// Try to parse as JSON first
	suggestions, err := s.parseJSON(response)
	if err == nil && len(suggestions) > 0 {
		return suggestions, nil
	}

	log.Printf("WARNING: Failed to parse as JSON, attempting fallback parsing: %v", err)

	// Fallback: Parse structured text
	suggestions = s.parseFallback(response)
	if len(suggestions) > 0 {
		return suggestions, nil
	}

	return nil, fmt.Errorf("could not parse any suggestions from response")
}

// parseJSON attempts to extract JSON array from the response
func (s *LLMService) parseJSON(response string) ([]FeatureSuggestion, error) {
	// Try to find JSON array in the response
	jsonStart := strings.Index(response, "[")
	jsonEnd := strings.LastIndex(response, "]")

	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		return nil, fmt.Errorf("no JSON array found in response")
	}

	jsonStr := response[jsonStart : jsonEnd+1]

	var suggestions []FeatureSuggestion
	if err := json.Unmarshal([]byte(jsonStr), &suggestions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return suggestions, nil
}

// parseFallback attempts to parse structured text format
func (s *LLMService) parseFallback(response string) []FeatureSuggestion {
	var suggestions []FeatureSuggestion

	// Pattern: "1. Feature Name - Description"
	// or "Feature Name: Description"
	lines := strings.Split(response, "\n")

	// Regex patterns to match different formats
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`^\d+\.\s*([^-:]+)\s*[-:]\s*(.+)$`),
		regexp.MustCompile(`^[-*]\s*([^-:]+)\s*[-:]\s*(.+)$`),
		regexp.MustCompile(`^([A-Z][^-:]+)\s*[-:]\s*(.+)$`),
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		for _, pattern := range patterns {
			matches := pattern.FindStringSubmatch(line)
			if len(matches) >= 3 {
				name := strings.TrimSpace(matches[1])
				description := strings.TrimSpace(matches[2])

				if name != "" && description != "" {
					suggestions = append(suggestions, FeatureSuggestion{
						Name:        name,
						Description: description,
						Category:    "Backend", // Default category
						Priority:    "Medium",  // Default priority
					})
					break
				}
			}
		}

		// Limit to 5 suggestions
		if len(suggestions) >= 5 {
			break
		}
	}

	return suggestions
}
