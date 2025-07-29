package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/google/go-github/v56/github"
	"github.com/yourusername/github-app-boilerplate/github/client"
	"github.com/yourusername/github-app-boilerplate/models"
)

type WebhookHandler struct {
	githubClient *client.Client
}

func NewWebhookHandler(githubClient *client.Client) *WebhookHandler {
	return &WebhookHandler{
		githubClient: githubClient,
	}
}

func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Verify the webhook signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if err := h.githubClient.VerifyWebhookSignature(payload, signature); err != nil {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Get the event type
	eventType := r.Header.Get("X-GitHub-Event")
	if eventType == "" {
		http.Error(w, "No X-GitHub-Event header found", http.StatusBadRequest)
		return
	}

	// Process the event
	switch eventType {
	case "pull_request":
		h.handlePullRequestEvent(w, payload)
	case "installation":
		h.handleInstallationEvent(w, payload)
	default:
		log.Printf("Unhandled event type: %s", eventType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Event received but not processed"))
	}
}

func (h *WebhookHandler) handlePullRequestEvent(w http.ResponseWriter, payload []byte) {
	var event models.PullRequestEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		http.Error(w, "Error parsing pull request event", http.StatusBadRequest)
		return
	}

	log.Printf("Processing pull request %s: %s", event.Action, event.PullRequest.GetHTMLURL())

	// Here you can add your business logic for pull request events
	switch event.Action {
	case "opened", "reopened", "synchronize":
		// Handle PR opened/reopened/updated
		log.Printf("PR #%d %s by %s", event.Number, event.Action, event.Sender.GetLogin())
		log.Printf("Title: %s", event.PullRequest.GetTitle())
		log.Printf("URL: %s", event.PullRequest.GetHTMLURL())
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Pull request event processed"))
}

func (h *WebhookHandler) handleInstallationEvent(w http.ResponseWriter, payload []byte) {
	var event models.InstallationEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		http.Error(w, "Error parsing installation event", http.StatusBadRequest)
		return
	}

	log.Printf("Processing installation event: %s", event.Action)

	switch event.Action {
	case "created":
		log.Printf("App installed on %d repositories", len(event.Repositories))
		for _, repo := range event.Repositories {
			log.Printf("- %s", repo.GetFullName())
		}
	case "deleted":
		log.Printf("App uninstalled from account %d", event.Installation.GetAccount().GetID())
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Installation event processed"))
}
