package github

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v56/github"
	"github.com/yourusername/github-app-boilerplate/config"
)

type Client struct {
	appClient     *github.Client
	config       *config.Config
	privateKey   *rsa.PrivateKey
}

func NewClient(cfg *config.Config) (*Client, error) {
	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[DEBUG][PR_RELEASE] Creating new GitHub client with app ID: %d\n", cfg.GitHubAppID)
	}

	// Parse the private key
	block, _ := pem.Decode([]byte(cfg.GitHubPrivateKeyPEM))
	if block == nil {
		if os.Getenv("DEBUG") == "1" {
			fmt.Printf("[DEBUG][PR_RELEASE] Failed to parse private key: invalid PEM block\n")
		}
		return nil, fmt.Errorf("failed to parse private key: invalid PEM block")
	}

	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[DEBUG][PR_RELEASE] Successfully decoded PEM block\n")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		if os.Getenv("DEBUG") == "1" {
			fmt.Printf("[DEBUG][PR_RELEASE] Failed to parse private key: %v\n", err)
		}
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}
	
	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[DEBUG][PR_RELEASE] Successfully parsed private key\n")
	}

	transport := ghinstallation.NewAppsTransport(
		ghinstallation.WithPrivateKey(privateKey),
		cfg.GitHubAppID,
	)

	return &Client{
		appClient:     github.NewClient(transport.Client()),
		config:       cfg,
		privateKey:   privateKey,
	}, nil
}

// GetInstallationToken returns an installation token for the given installation ID
func (c *Client) GetInstallationToken(installationID int64) (string, error) {
	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[DEBUG][PR_RELEASE] Getting installation token for installation ID: %d\n", installationID)
	}

	tr := ghinstallation.NewFromAppsTransport(
		ghinstallation.NewAppsTransport(
			ghinstallation.WithPrivateKey(c.privateKey),
			c.config.GitHubAppID,
		),
		installationID,
	)

	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[DEBUG][PR_RELEASE] Requesting token from GitHub API\n")
	}

	token, err := tr.Token(context.Background())
	if err != nil {
		if os.Getenv("DEBUG") == "1" {
			fmt.Printf("[DEBUG][PR_RELEASE] Failed to get installation token: %v\n", err)
		}
		return "", fmt.Errorf("failed to get installation token: %v", err)
	}

	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[DEBUG][PR_RELEASE] Successfully generated installation token for installation ID %d\n", installationID)
	}

	log.Printf("Successfully generated installation token for installation ID %d", installationID)
	return token, nil
}

// NewInstallationClient creates a new GitHub client authenticated as an installation
func (c *Client) NewInstallationClient(installationID int64) (*github.Client, error) {
	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[DEBUG][PR_RELEASE] Creating new installation client for installation ID: %d\n", installationID)
	}

	tr := ghinstallation.NewFromAppsTransport(
		ghinstallation.NewAppsTransport(
			ghinstallation.WithPrivateKey(c.privateKey),
			c.config.GitHubAppID,
		),
		installationID,
	)

	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[DEBUG][PR_RELEASE] Successfully created GitHub client for installation ID: %d\n", installationID)
	}

	return github.NewClient(tr.Client()), nil
}

// VerifyWebhookSignature verifies the signature of a webhook payload
func (c *Client) VerifyWebhookSignature(payload []byte, signature string) error {
	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[DEBUG][PR_RELEASE] Verifying webhook signature\n")
	}

	err := ghinstallation.VerifyPayload(payload, signature, []byte(c.config.GitHubWebhookSecret))
	if err != nil {
		if os.Getenv("DEBUG") == "1" {
			fmt.Printf("[DEBUG][PR_RELEASE] Webhook signature verification failed: %v\n", err)
		}
		return err
	}

	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[DEBUG][PR_RELEASE] Webhook signature verified successfully\n")
	}

	return nil
}
