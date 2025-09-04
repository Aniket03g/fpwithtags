package tests

import (
	"fmt"
	"net/http"
	"os"

	"github.com/FeaturePlus/pkg/featureplus"
)

func TestClient() {
	// Create client with direct token
	client := featureplus.NewClient("http://localhost:8080", &http.Client{})
	
	// Set token directly
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTcwMDY4NTEsInVzZXJfaWQiOjN9.J7pfi_mkynpcCSnVCdRuoWRuyC-7vJnMRCaRLmfMoyo"
	client.SetAuthToken(token)
	
	fmt.Printf("Using token: %s...\n", token[:20])
	
	// Test API call
	fmt.Println("Making API call to list PRs...")
	prs, err := client.ListPRs(0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Success! Found %d PRs\n", len(prs))
}
