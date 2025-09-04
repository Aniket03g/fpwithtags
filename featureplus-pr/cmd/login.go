package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"
	"featureplus-pr/internal/config"
)

var (
	username string
	password string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to FeaturePlus API",
	Long:  `Login to FeaturePlus API using your username/email and password.`,
	Run: func(cmd *cobra.Command, args []string) {
		if username == "" || password == "" {
			fmt.Println("Error: Username/email and password are required")
			cmd.Help()
			return
		}

		// Prepare login request
		loginURL := config.GetAPIURL() + "/api/auth/login"
		requestBody, err := json.Marshal(map[string]string{
			"email":    username, // The backend accepts email field for login
			"password": password,
		})
		if err != nil {
			fmt.Printf("Error preparing login request: %v\n", err)
			return
		}

		// Send login request
		resp, err := http.Post(loginURL, "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			fmt.Printf("Error connecting to server: %v\n", err)
			return
		}
		defer resp.Body.Close()

		// Read response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			return
		}

		// Check response status
		if resp.StatusCode != http.StatusOK {
			var errorResp struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error != "" {
				fmt.Printf("Login failed: %s\n", errorResp.Error)
			} else {
				fmt.Printf("Login failed with status code: %d\n", resp.StatusCode)
			}
			return
		}

		// Parse successful response
		var loginResp struct {
			Token    string `json:"token"`
			AuthInfo struct {
				ID       float64 `json:"id"`
				Username string  `json:"username"`
				Email    string  `json:"email"`
				Role     string  `json:"role"`
			} `json:"auth_info"`
		}

		if err := json.Unmarshal(body, &loginResp); err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			return
		}

		// Save authentication info
		userID := uint(loginResp.AuthInfo.ID)
		if err := config.SaveAuthInfo(
			loginResp.Token,
			userID,
			loginResp.AuthInfo.Username,
			loginResp.AuthInfo.Email,
			loginResp.AuthInfo.Role,
		); err != nil {
			fmt.Printf("Error saving authentication info: %v\n", err)
			return
		}

		fmt.Printf("Successfully logged in as %s (%s)\n", loginResp.AuthInfo.Username, loginResp.AuthInfo.Role)
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	// Add flags for username and password
	loginCmd.Flags().StringVarP(&username, "username", "u", "", "Username or email for login")
	loginCmd.Flags().StringVarP(&password, "password", "p", "", "Password for login")

	// Mark flags as required
	loginCmd.MarkFlagRequired("username")
	loginCmd.MarkFlagRequired("password")
}
