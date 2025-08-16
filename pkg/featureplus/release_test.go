package featureplus

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateRelease(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/releases" {
			t.Errorf("Expected to request '/api/releases', got: %s", r.URL.Path)
		}
		
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got: %s", r.Method)
		}
		
		// Decode request body
		var req CreateReleaseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Error decoding request body: %v", err)
		}
		
		// Verify request fields
		if req.Tag != "v1.0.0" {
			t.Errorf("Expected tag 'v1.0.0', got: %s", req.Tag)
		}
		if req.Notes != "Test release" {
			t.Errorf("Expected notes 'Test release', got: %s", req.Notes)
		}
		if len(req.PRIDs) != 2 || req.PRIDs[0] != 1 || req.PRIDs[1] != 2 {
			t.Errorf("Expected PR IDs [1, 2], got: %v", req.PRIDs)
		}
		
		// Return mock response
		release := Release{
			ID:        1,
			Tag:       req.Tag,
			Notes:     req.Notes,
			Status:    "Created",
			CreatedAt: 1629123456,
			PRIDs:     req.PRIDs,
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()
	
	// Create client with test server URL
	client := NewClient(server.URL, nil)
	
	// Create release request
	req := &CreateReleaseRequest{
		Tag:   "v1.0.0",
		Notes: "Test release",
		PRIDs: []uint{1, 2},
	}
	
	// Call the function
	release, err := client.CreateRelease(req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Verify response
	if release.ID != 1 {
		t.Errorf("Expected release ID 1, got: %d", release.ID)
	}
	if release.Tag != "v1.0.0" {
		t.Errorf("Expected release tag 'v1.0.0', got: %s", release.Tag)
	}
	if release.Status != "Created" {
		t.Errorf("Expected release status 'Created', got: %s", release.Status)
	}
}

func TestListReleases(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/releases" {
			t.Errorf("Expected to request '/api/releases', got: %s", r.URL.Path)
		}
		
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got: %s", r.Method)
		}
		
		// Return mock response
		releases := []Release{
			{
				ID:        1,
				Tag:       "v1.0.0",
				Notes:     "First release",
				Status:    "Created",
				CreatedAt: 1629123456,
				PRIDs:     []uint{1, 2},
			},
			{
				ID:        2,
				Tag:       "v1.1.0",
				Notes:     "Second release",
				Status:    "Finalized",
				CreatedAt: 1629223456,
				PRIDs:     []uint{3, 4},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(releases)
	}))
	defer server.Close()
	
	// Create client with test server URL
	client := NewClient(server.URL, nil)
	
	// Call the function
	releases, err := client.ListReleases()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Verify response
	if len(releases) != 2 {
		t.Fatalf("Expected 2 releases, got: %d", len(releases))
	}
	
	if releases[0].ID != 1 || releases[0].Tag != "v1.0.0" {
		t.Errorf("First release incorrect: %+v", releases[0])
	}
	
	if releases[1].ID != 2 || releases[1].Tag != "v1.1.0" {
		t.Errorf("Second release incorrect: %+v", releases[1])
	}
}

func TestFinalizeRelease(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/releases/1/finalize" {
			t.Errorf("Expected to request '/api/releases/1/finalize', got: %s", r.URL.Path)
		}
		
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got: %s", r.Method)
		}
		
		// Return success
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	// Create client with test server URL
	client := NewClient(server.URL, nil)
	
	// Call the function
	err := client.FinalizeRelease(1)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}
