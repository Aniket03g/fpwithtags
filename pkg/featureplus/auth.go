package featureplus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// AuthResponse represents the response from the login API
type AuthResponse struct {
	Token    string `json:"token"`
	AuthInfo struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	} `json:"auth_info"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login authenticates with the FeaturePlus API and returns the auth response
func (c *Client) Login(email, password string) (*AuthResponse, error) {
	url := fmt.Sprintf("%s/api/auth/login", c.BaseURL)

	reqBody, err := json.Marshal(LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("error marshaling login request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil && errorResp.Error != "" {
			return nil, fmt.Errorf("login failed: %s", errorResp.Error)
		}
		return nil, fmt.Errorf("login failed with status code: %d", resp.StatusCode)
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &authResp, nil
}

// SetAuthToken sets the authentication token for the client
func (c *Client) SetAuthToken(token string) {
	c.authToken = token
}

// GetAuthToken returns the current authentication token
func (c *Client) GetAuthToken() string {
	return c.authToken
}
