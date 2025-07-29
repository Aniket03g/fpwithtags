package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yourusername/github-app-boilerplate/handlers"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(webhookHandler *handlers.WebhookHandler) *mux.Router {
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// GitHub webhook endpoint
	api.HandleFunc("/github/webhooks", webhookHandler.HandleWebhook).Methods("POST")

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json"
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}).Methods("GET")

	// Add request logging middleware
	r.Use(loggingMiddleware)

	return r
}

// loggingMiddleware logs the HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request
		log.Printf("[%s] %s %s %s", r.RemoteAddr, r.Method, r.URL.Path, r.URL.RawQuery)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// Helper function to log messages
func log(message string) {
	// You can replace this with your preferred logging mechanism
	println(message)
}
