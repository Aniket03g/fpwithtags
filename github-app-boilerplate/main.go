package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/github-app-boilerplate/config"
	"github.com/yourusername/github-app-boilerplate/github"
	"github.com/yourusername/github-app-boilerplate/handlers"
	"github.com/yourusername/github-app-boilerplate/routes"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize GitHub client
	githubClient, err := github.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create GitHub client: %v", err)
	}

	// Initialize webhook handler
	webhookHandler := handlers.NewWebhookHandler(githubClient)

	// Set up routes
	router := routes.SetupRoutes(webhookHandler)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s...\n", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start server: %v\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}

	log.Println("Server stopped")
}
